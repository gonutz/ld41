// Package mixer provides an abstraction over the sound card to be able to play
// multiple sounds simultaneously, combining different effects.
//
// Call Init to start the mixer and Close when you are done with it.
// Call NewSoundSource to create a sound source from PCM data. You can use this
// source to play the sound using the Play... functions. Each call will give you
// a Sound which represents that particular instance of the sound source that
// is then played. You can change its parameters, pause and resume it and change
// its position. Once the Sound is done playing, its Stopped function will
// return true. In this case you cannot use the Sound anymore and should set its
// pointer to nil so the Go garbage collector can remove it. Calling any
// function on a Stopped Sound has no effect.
package mixer

import (
	"sync"
	"time"

	"github.com/gonutz/mixer/dsound"
)

// TODO right now the volume and pan only change in discrete chunks, whenever
// the sound buffer is written. To be accurate, individual samples would have
// to be modified depeding on when they are played, this can get hairy to
// implement so think about if it makes sense to do that or if the current
// solution is good enough
// TODO have SetPitch in Sound? Or In SoundSource?

var (
	// writeCursor keeps the offset into DirectSound's ring buffer at which data
	// was written last
	writeCursor uint

	// sounds are all currently active sounds (they might be paused) from which
	// the mixed sound output is computed
	sounds []*sound

	// lock is for changes to the mixer state and changes to the sound, these
	// must not occur while mixing sound data
	lock sync.Mutex

	// volume is in the range from 0 (silent) to 1 (full volume)
	volume float32

	// stop is a signalling channel for the mixer to know when to stop updating
	// (which happens in a separate Go routine), e.g. after Close was called or
	// when an error happened from which it cannot recover
	stop chan bool

	// writeAheadBuffer and mixBuffer are the buffers for mixing the sound
	// sources; their size determines the time of the sound that will be output
	// for future playing in every mixer update
	writeAheadBuffer []byte
	mixBuffer        []float32
	// leftBuffer and rightBuffer are simply pointers into the mixBuffer's first
	// and second half
	leftBuffer, rightBuffer []float32

	// lastError keeps the last error encountered by the mixer; it can be
	// queried by the client using the Error function
	lastError error

	// inited is used to coordinate multiple and/or concurrent calls to Init
	// and Close
	inited   bool
	initLock sync.Mutex
)

const (
	bytesPerSample = 4                      // 2 channels, 16 bit each
	bytesPerSecond = 44100 * bytesPerSample // fixed sample frequency of 44100Hz
)

// Init sets up DirectSound and prepares for mixing and playing sounds. It
// starts a Go routine that periodically writes to the sound buffer to output
// to the sound card.
// Call Close when you are done with the mixer.
func Init() error {
	initLock.Lock()
	defer initLock.Unlock()
	if inited {
		return nil
	}

	if err := dsound.Init(44100); err != nil {
		return err
	}

	writeAheadByteCount := bytesPerSecond / 10 // buffer 100ms
	// make sure it is evenly dividable into samples
	writeAheadByteCount -= writeAheadByteCount % bytesPerSample
	writeAheadBuffer = make([]byte, writeAheadByteCount)
	mixBuffer = make([]float32, writeAheadByteCount/2) // 2 bytes form one value
	leftBuffer = mixBuffer[:len(mixBuffer)/2]
	rightBuffer = mixBuffer[len(mixBuffer)/2:]
	volume = 1

	// initially write silence to sound buffer
	if err := dsound.WriteToSoundBuffer(writeAheadBuffer, 0); err != nil {
		return err
	}
	if err := dsound.StartSound(); err != nil {
		return err
	}

	stop = make(chan bool)
	go func() {
		pulse := time.Tick(10 * time.Millisecond)
		for {
			select {
			case <-pulse:
				update()
				if lastError != nil {
					return
				}
			case <-stop:
				return
			default:
				time.Sleep(1 * time.Millisecond)
			}
		}
	}()

	inited = true

	return nil
}

// Close blocks until playing sound is stopped. It shuts down the DirectSound
// system.
func Close() {
	initLock.Lock()
	defer initLock.Unlock()
	if !inited {
		return
	}

	stop <- true
	dsound.StopSound()
	dsound.Close()

	inited = false
}

// Error returns the last error that occurred. If a fatal error occurs, the Go
// routine for mixing and playing sounds might stop before you call Close. In
// this case, call Error to retrieve the cause of the failure.
func Error() error {
	return lastError
}

// SetVolume sets the master volume. All sounds will be scaled by this factor.
// It is in the range [0..1] and will be clamped to it.
func SetVolume(v float32) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	lock.Lock()
	defer lock.Unlock()

	volume = v
}

func update() {
	lock.Lock()
	defer lock.Unlock()

	_, write, err := dsound.GetPlayAndWriteCursors()
	if err != nil {
		lastError = err
		return
	}
	if write != writeCursor {
		var delta uint
		if write > writeCursor {
			delta = write - writeCursor
		} else {
			// wrap-around happened in DirectSound's ring buffer
			delta = write + dsound.BufferSize() - writeCursor
		}
		advanceSoundsByBytes(int(delta))

		// rewrite the whole look-ahead with newly mixed data
		lastError = dsound.WriteToSoundBuffer(mix(), write)
		if lastError != nil {
			return
		}
	}
	writeCursor = write
}

func mix() []byte {
	for i := range mixBuffer {
		mixBuffer[i] = 0.0
	}

	for _, sound := range sounds {
		sound.addToMixBuffer()
	}

	out := 0
	for i := range leftBuffer {
		writeAheadBuffer[out], writeAheadBuffer[out+1] = floatToBytes(leftBuffer[i] * volume)
		writeAheadBuffer[out+2], writeAheadBuffer[out+3] = floatToBytes(rightBuffer[i] * volume)
		out += 4
	}

	return writeAheadBuffer
}

func floatToBytes(f float32) (lo, hi byte) {
	if f < 0 {
		if f < -1 {
			f = -1
		}
		value := int16(f * 32768)
		return byte(value & 0xFF), byte((value >> 8) & 0xFF)
	}

	if f > 1 {
		f = 1
	}
	value := int16(f * 32767)
	return byte(value & 0xFF), byte((value >> 8) & 0xFF)

}

func advanceSoundsByBytes(byteCount int) {
	for i := 0; i < len(sounds); i++ {
		if !sounds[i].paused {
			sounds[i].advanceBySamples(byteCount / 4)
			if sounds[i].isOver() {
				sounds[i].source = nil
				sounds = append(sounds[:i], sounds[i+1:]...)
				i--
			}
		}
	}
}

func byteToFloat(b byte) float32 {
	// for 8 bit sound data, the value 128 is silence.
	if b >= 128 {
		return float32(b-128) / 127.0
	}
	return (float32(b) - 128) / 128.0
}

func makeFloat(b1, b2 byte) float32 {
	// 16 bit sound data is in little endian byte order signed int16s
	lo, hi := uint16(b1), uint16(b2)
	val := int16(lo | (hi << 8))
	if val >= 0 {
		return float32(val) / 32767
	}
	return float32(val) / 32768
}
