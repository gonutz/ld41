package main

import "github.com/gonutz/prototype/draw"

type playingState struct{}

func (*playingState) enter(state) {}
func (*playingState) leave()      {}

func (*playingState) update(window draw.Window) state {
	if window.WasKeyPressed(draw.KeyEscape) {
		window.Close()
	}
	window.DrawScaledText("TODO: game", 10, 10, 3, draw.White)
	return playing
}
