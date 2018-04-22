package main

import (
	"github.com/gonutz/prototype/draw"
	"math/rand"
	"time"
)

const (
	windowTitle      = "No-Brain Jogging"
	windowW, windowH = 1000, 600
	musicLength      = 8081 * time.Millisecond
)

type state interface {
	enter(from state)
	update(window draw.Window) state
	leave()
}

// all game states
var (
	loading      = &loadingState{}
	menu         = &menuState{}
	playing      = &playingState{}
	dead         = &deadState{}
	instructions = &instructionsState{}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var state state = loading
	state.enter(nil)

	defer cleanUpAssets()

	var musicStart time.Time

	check(draw.RunWindow(windowTitle, windowW, windowH, func(window draw.Window) {
		newState := state.update(window)
		if state != newState {
			state.leave()
			newState.enter(state)
		}
		state = newState

		if loading.assetsLoaded {
			now := time.Now()
			if now.Sub(musicStart) >= musicLength {
				window.PlaySoundFile(file("music.wav"))
				musicStart = now
			}
		}
	}))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
