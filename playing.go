package main

import "github.com/gonutz/prototype/draw"

const (
	playerSpeed      = 4
	playerW, playerH = 93, 209
)

type playingState struct {
	playerX          int
	playerFacingLeft bool
}

func (s *playingState) enter(state) {
	s.playerX = windowW / 4
}

func (*playingState) leave() {}

func (s *playingState) update(window draw.Window) state {
	// handle input
	if window.WasKeyPressed(draw.KeyEscape) {
		// TODO eventually go to a pause menu here
		window.Close()
	}
	if window.IsKeyDown(draw.KeyLeft) {
		s.playerX -= playerSpeed
		s.playerFacingLeft = true
	}
	if window.IsKeyDown(draw.KeyRight) {
		s.playerX += playerSpeed
		s.playerFacingLeft = false
	}

	// render
	window.FillRect(0, 0, windowW, windowH, draw.RGB(0, 0.5, 1))
	if s.playerFacingLeft {
		window.DrawImageFile("rsc/hero left.png", s.playerX, 250)
	} else {
		window.DrawImageFile("rsc/hero right.png", s.playerX, 250)
	}

	return playing
}
