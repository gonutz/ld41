package main

import (
	"github.com/gonutz/prototype/draw"
	"math/rand"
	"time"
)

const (
	title            = "LD41 - TODO: name this game"
	windowW, windowH = 1000, 600
)

type state interface {
	enter(from state)
	update(window draw.Window) state
	leave()
}

// all game states
var (
	playing state = &playingState{}
)

func main() {
	rand.Seed(time.Now().UnixNano())
	state := playing
	state.enter(nil)
	check(draw.RunWindow(title, windowW, windowH, func(window draw.Window) {
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
