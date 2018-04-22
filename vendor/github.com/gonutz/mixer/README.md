mixer
=====

See [the documentation](https://godoc.org/github.com/gonutz/mixer) for an API overview.

Your main interface to use is the `SoundSource`:

```Go
type SoundSource interface {
    // PlayPaused adds a new one-time sound to the mixer. It is in paused state.
    PlayPaused() Sound
    // PlayOnce adds a new one-time sound to the mixer. It is started right away
    // and stopped when it finishes.
    PlayOnce() Sound

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
```