package mixer

import "time"

type Sound interface {
	// SetPaused starts or stops the sound. Note that the sound position is not
	// changed with this function, meaning that if the sound is not playing
	// because it was played all the way to the end, calling SetPaused(false)
	// will not restart it from the beginning. You have to call SetPosition(0)
	// to reset the sound to the start. If the sound is not paused, it will then
	// play right away.
	SetPaused(bool)

	// Paused returns the last value set in SetPaused. It does not consider
	// whether the sound is being played right now. It may not be paused but
	// could have reached the end and thus is not audible although not paused.
	Paused() bool

	// Playing returns true if the sound is not paused and has not reached the
	// end.
	Playing() bool

	// Stopped returns true if the sound has been fully played. This means that
	// the user cannot use the sound anymore. Set the pointer to nil in this
	// case so that the Go runtime can free its memory on the next GC.
	Stopped() bool

	// SetVolume sets the volume factor for all channels. Its range is [0..1]
	// and it will be clamped to that range.
	// Note that the audible difference in loudness between 100% and 50% is the
	// same as between 50% and 25% and so on. Changing the sound on a
	// logarithmic scale will sound to the human ear as if you decrease the
	// sound by equal steps.
	SetVolume(float32)

	// Volume returns a value in the range of 0 (silent) to 1 (full volume).
	Volume() float32

	// SetPan changes the volume ratio between left and right output channel.
	// Setting it to -1 will make channel 1 (left speaker) output at 100% volume
	// while channel 2 (right speaker) has a volume of 0%.
	// A pan of 0 means both speakers' volumes are at 100%, +1 means the left
	// speaker is silenced.
	// This value is clamped to [-1..1]
	SetPan(float32)

	// Pan returns the current pan as a value in the range of -1 (only left
	// speaker) to 1 (only right speaker). A value of 0 means both speakers play
	// at full volume.
	Pan() float32

	// Length is the length of the whole sound, it does not consider how far it
	// is already played or if it loops or not.
	Length() time.Duration

	// SetPosition sets the time offset into the sound at which it will continue
	// to play.
	SetPosition(time.Duration)

	// Position is the current offset from the start of the sound. It changes
	// while the sound is played.
	Position() time.Duration
}

type sound struct {
	source                        *soundSource
	cursor                        int
	paused                        bool
	volume                        float32
	pan                           float32
	leftPanFactor, rightPanFactor float32
}

func (s *sound) SetPaused(paused bool) {
	if s.source == nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	s.paused = paused
}

func (s *sound) Paused() bool {
	return s.paused
}

func (s *sound) Playing() bool {
	return !s.paused && s.source != nil && s.cursor < len(s.source.left)
}

func (s *sound) Stopped() bool {
	return s.source == nil
}

func (s *sound) SetVolume(v float32) {
	if s.source == nil {
		return
	}

	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}

	lock.Lock()
	defer lock.Unlock()

	s.volume = v
}

func (s *sound) Volume() float32 {
	return s.volume
}

func (s *sound) SetPan(p float32) {
	if s.source == nil {
		return
	}

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

	lock.Lock()
	defer lock.Unlock()

	s.pan = p
	s.leftPanFactor, s.rightPanFactor = left, right
}

func (s *sound) Pan() float32 {
	return float32(s.pan)
}

func (s *sound) Length() time.Duration {
	if s.source == nil {
		return 0
	}
	return s.source.Length() // TODO * loops
}

func (s *sound) SetPosition(pos time.Duration) {
	if s.source == nil {
		return
	}

	lock.Lock()
	defer lock.Unlock()

	s.cursor = int(bytesPerSecond*pos.Seconds()/4.0 + 0.5)
	if s.cursor < 0 {
		s.cursor = 0
	}
	if s.cursor > len(s.source.left) {
		s.cursor = len(s.source.left)
	}
}

func (s *sound) Position() time.Duration {
	return time.Duration(float64(s.cursor)/bytesPerSecond*4000000000) * time.Nanosecond
}

func (s *sound) advanceBySamples(sampleCount int) {
	s.cursor += sampleCount
	if s.cursor > len(s.source.left) {
		s.cursor = len(s.source.left)
	}
}

func (s *sound) addToMixBuffer() {
	if s.paused {
		return
	}

	writeTo := s.cursor + len(leftBuffer)
	if writeTo > len(s.source.left) {
		writeTo = len(s.source.left)
	}

	leftFactor := s.volume * s.leftPanFactor
	rightFactor := s.volume * s.rightPanFactor
	out := 0
	for i := s.cursor; i < writeTo; i++ {
		leftBuffer[out] += s.source.left[i] * leftFactor
		rightBuffer[out] += s.source.right[i] * rightFactor
		out++
	}
}

func (s *sound) isOver() bool {
	// TODO consider loops
	return s.cursor >= len(s.source.left)
}
