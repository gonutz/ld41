package mixer

import (
	"fmt"
	"time"

	"github.com/gonutz/mixer/wav"
)

type SoundSource interface {
	// PlayPaused adds a new one-time sound to the mixer. It is in paused state.
	PlayPaused() Sound
	// PlayOnce adds a new one-time sound to the mixer. It is started right away
	// and stopped when it finishes.
	PlayOnce() Sound

	//PlayLooping(loops int) Sound
	//PlayForeverLooping() Sound

	// SetVolume sets the default volume for all sounds played in the future.
	// Changing the Sound's volume will simply overwrite this setting (instead
	// of combining the factors).
	// The range is [0..1] and it is clamped to that.
	SetVolume(float32)
	Volume() float32

	// SetPan sets the default pan for all sounds played in the future.
	// Changing the Sound's pan will simply overwrite this setting (instead
	// of combining the factors).
	SetPan(float32)
	Pan() float32

	// Length returns the duration of the sound data. Note that a played Sound
	// may have a different value for its Length function as it considers
	// looping.
	Length() time.Duration
}

// NewSound creates a new sound source from the given wave data and starts
// playing it right away. You can call SetPlaying(false) on the returned sound
// if you do not want to play the sound right away.
func NewSoundSource(w *wav.Wave) (SoundSource, error) {
	left, right, err := makeTwoChannelFloats(w)
	if err != nil {
		return nil, err
	}

	source := &soundSource{
		left:           left,
		right:          right,
		volume:         1,
		pan:            0,
		leftPanFactor:  1,
		rightPanFactor: 1,
	}

	return source, nil
}

// TODO resample the right frequency
func makeTwoChannelFloats(w *wav.Wave) (left, right []float32, err error) {
	if w.ChannelCount == 1 && w.BitsPerSample == 8 {
		result := make([]float32, len(w.Data))
		left = result
		right = result
		for i := range w.Data {
			result[i] = byteToFloat(w.Data[i])
		}
	} else if w.ChannelCount == 1 && w.BitsPerSample == 16 {
		result := make([]float32, len(w.Data)/2)
		left = result
		right = result
		in := 0
		for i := range result {
			result[i] = makeFloat(w.Data[in], w.Data[in+1])
			in += 2
		}
	} else if w.ChannelCount == 2 && w.BitsPerSample == 8 {
		result := make([]float32, len(w.Data))
		left, right = result[:len(result)/2], result[len(result)/2:]
		in := 0
		for i := range left {
			left[i] = byteToFloat(w.Data[in])
			right[i] = byteToFloat(w.Data[in+1])
			in += 2
		}
	} else if w.ChannelCount == 2 && w.BitsPerSample == 16 {
		data := w.Data
		if len(data)%4 != 0 {
			data = data[:len(data)-len(data)%4]
		}
		result := make([]float32, len(data)/2)
		left, right = result[:len(result)/2], result[len(result)/2:]
		in := 0
		for i := range left {
			left[i] = makeFloat(data[in], data[in+1])
			right[i] = makeFloat(data[in+2], data[in+3])
			in += 4
		}
	} else {
		return nil, nil, fmt.Errorf(
			"mixer.NewSoundSource: unsupported format: "+
				"%v channels (must be 1 or 2), "+
				"%v bits per sample (must be 8 or 16)",
			w.ChannelCount, w.BitsPerSample)
	}
	return
}

type soundSource struct {
	left, right                   []float32
	volume                        float32
	pan                           float32
	leftPanFactor, rightPanFactor float32
}

func (s *soundSource) PlayOnce() Sound {
	return s.play(false)
}

func (s *soundSource) PlayPaused() Sound {
	return s.play(true)
}

func (s *soundSource) play(paused bool) Sound {
	sound := &sound{
		source:         s,
		paused:         paused,
		volume:         s.volume,
		pan:            s.pan,
		leftPanFactor:  s.leftPanFactor,
		rightPanFactor: s.rightPanFactor,
	}

	lock.Lock()
	defer lock.Unlock()

	sounds = append(sounds, sound)
	return sound
}

func (s *soundSource) SetVolume(v float32) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	s.volume = v
}

func (s *soundSource) Volume() float32 {
	return s.volume
}

func (s *soundSource) SetPan(p float32) {
	if p < -1 {
		p = -1
	}
	if p > 1 {
		p = 1
	}

	left, right := float32(1), float32(1)
	if p < 0 {
		right = 1 + p
	}
	if p > 0 {
		left = 1 - p
	}

	s.pan = p
	s.leftPanFactor, s.rightPanFactor = left, right
}

func (s *soundSource) Pan() float32 {
	return s.pan
}

func (s *soundSource) Length() time.Duration {
	return time.Duration(float64(len(s.left))/bytesPerSecond*4000000000) * time.Nanosecond
}
