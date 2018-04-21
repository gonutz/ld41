package main

import (
	"fmt"
	"github.com/gonutz/prototype/draw"
	"math/rand"
)

const (
	playerSpeed      = 4
	playerW, playerH = 93, 209
)

type playingState struct {
	playerX          int
	playerFacingLeft bool
	generator        mathGenerator
	assignment       assignment
}

func (s *playingState) enter(state) {
	s.playerX = windowW / 4
	s.generator = mathGenerator{
		ops: []mathOp{add, subtract, add, subtract, multiply, divide},
		max: 9,
	}
	s.assignment = s.generator.generate(rand.Int)
}

func (*playingState) leave() {}

func (s *playingState) update(window draw.Window) state {
	// handle input
	if window.WasKeyPressed(draw.KeyEscape) {
		// TODO eventually go to a pause menu here
		window.Close()
	}
	if window.WasKeyPressed(draw.KeyEnter) {
		s.assignment = s.generator.generate(rand.Int)
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
	const playerY = 250
	window.FillRect(0, 0, windowW, windowH, draw.RGB(0, 0.5, 1))
	if s.playerFacingLeft {
		window.DrawImageFile("rsc/hero left.png", s.playerX, playerY)
	} else {
		window.DrawImageFile("rsc/hero right.png", s.playerX, playerY)
	}
	const mathScale = 2
	w, h := window.GetScaledTextSize(s.assignment.question, mathScale)
	window.DrawScaledText(
		s.assignment.question,
		s.playerX+playerW/2-w/2,
		playerY-h,
		mathScale,
		draw.White,
	)
	window.DrawScaledText(fmt.Sprintf("= %d", s.assignment.answer), 0, 0, mathScale, draw.White)

	return playing
}
