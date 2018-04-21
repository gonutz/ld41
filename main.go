package main

import (
	"github.com/gonutz/prototype/draw"
	"math/rand"
	"time"
)

const (
	windowTitle      = "Shootematics"
	windowW, windowH = 1000, 600
)

type state interface {
	enter(from state)
	update(window draw.Window) state
	leave()
}

// all game states
var (
	loading = &loadingState{}
	menu    = &menuState{}
	playing = &playingState{}
	dead    = &deadState{}
)

func main() {
	rand.Seed(time.Now().UnixNano())

	var state state = loading
	state.enter(nil)

	defer cleanUpAssets()

	check(draw.RunWindow(windowTitle, windowW, windowH, func(window draw.Window) {
		newState := state.update(window)
		if state != newState {
			state.leave()
			newState.enter(state)
		}
		state = newState
	}))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
