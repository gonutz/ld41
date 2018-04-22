package dsound

import (
	"errors"
	"strconv"

	"github.com/gonutz/ds"
	"github.com/gonutz/w32"
)

var (
	globalDirectSoundObject  *ds.DirectSound
	globalPrimarySoundBuffer *ds.Buffer
	globalSoundBuffer        *ds.Buffer
	globalBufferSize         uint32
)

// Init sets up DirectSound and creates a sound buffer with 2 channels, 16 bit
// samples and the given sample frequency. The buffer is not played until you
// call StartSound.
// Make sure to call Close when you are done with DirectSound.
func Init(samplesPerSecond int) error {
	if samplesPerSecond <= 0 {
		return errors.New(
			"initDirectSound: illegal samplesPerSound: " +
				strconv.Itoa(samplesPerSecond))
	}

	return initDirectSound(samplesPerSecond)
}

func initDirectSound(samplesPerSecond int) error {
	dsound, err := ds.Create(nil)
	if err != nil {
		return err
	}
	err = dsound.SetCooperativeLevel(ds.HWND(w32.GetDesktopWindow()), ds.SCL_PRIORITY)
	if err != nil {
		dsound.Release()
		return err
	}
	primaryBuffer, err := dsound.CreateSoundBuffer(ds.BUFFERDESC{
		Flags:       ds.BCAPS_PRIMARYBUFFER,
		BufferBytes: 0, // NOTE must be 0 for primary buffer
	})
	if err != nil {
		dsound.Release()
		return err
	}
	format := ds.WAVEFORMATEX{
		FormatTag:     ds.WAVE_FORMAT_PCM,
		Channels:      2,
		SamplesPerSec: 44100,
		BitsPerSample: 16, // NOTE must be 8 or 16
	}
	format.BlockAlign = (format.Channels * format.BitsPerSample) / 8
	format.AvgBytesPerSec = format.SamplesPerSec * uint32(format.BlockAlign)
	err = primaryBuffer.SetFormat(format)
	if err != nil {
		primaryBuffer.Release()
		dsound.Release()
		return err
	}
	globalBufferSize = 2 * format.AvgBytesPerSec
	secondaryBuffer, err := dsound.CreateSoundBuffer(ds.BUFFERDESC{
		Flags:       ds.BCAPS_GETCURRENTPOSITION2 | ds.BCAPS_GLOBALFOCUS,
		BufferBytes: globalBufferSize,
		WfxFormat:   &format,
	})
	if err != nil {
		primaryBuffer.Release()
		dsound.Release()
		return err
	}

	globalDirectSoundObject = dsound
	globalPrimarySoundBuffer = primaryBuffer
	globalSoundBuffer = secondaryBuffer

	return nil
}

// Close releases all resources that were allocated when initializing
// DirectSound. It will stop playing the sound, if any.
func Close() {
	globalBufferSize = 0
	globalSoundBuffer.Release()
	globalPrimarySoundBuffer.Release()
	globalDirectSoundObject.Release()
}

// BufferSize returns the size in bytes of the sound buffer that you write to
// with WriteToSoundBuffer. When DirectSound is not initialized this value is 0.
func BufferSize() uint {
	return uint(globalBufferSize)
}

// StartSound must be called after initialization to make the sound buffer
// audible. It will internally call Play on the DirectSound buffer with the
// looping option so the sound plays forever (until you call StopSound).
func StartSound() error {
	return globalSoundBuffer.Play(0, ds.BPLAY_LOOPING)
}

// StopSound stops playing the sound buffer.
func StopSound() error {
	return globalSoundBuffer.Stop()
}

// WriteToSoundBuffer locks the sound buffer and writes the given data into it,
// starting at the given byte offset. The buffer is a ring buffer so writing
// outside the bounds will wrap around and continue writing to the beginning.
func WriteToSoundBuffer(data []byte, offset uint) error {
	mem, err := globalSoundBuffer.Lock(uint32(offset), uint32(len(data)), 0)
	if err != nil {
		return err
	}
	mem.Write(0, data)
	return globalSoundBuffer.Unlock(mem)
}

// GetPlayAndWriteCursors returns the play and write cursors. These are byte
// offsets into the sound buffer. The range between the two is commited to the
// sound card for playing so it is not safe to write into that area. According
// to the DirectSound documentation this area is usually about 15ms worth of
// data but a test on Windows 8.1 showed a value of 30ms.
// Note that the sound buffer is a ring buffer which is why the play cursor,
// which indicates the start of the commited region, can be a higher value than
// the write cursor, which indicates the end of the commited region.
// In the non-border case the play cursor will be less than the write cursor.
// You can safely  write sound data starting at the write cursor and ending at
// the play cursor.
func GetPlayAndWriteCursors() (play, write uint, err error) {
	p, w, e := globalSoundBuffer.GetCurrentPosition()
	play = uint(p)
	write = uint(w)
	err = e
	return
}
