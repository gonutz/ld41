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

var fireKeys = [10][2]draw.Key{
	{draw.Key0, draw.KeyNum0},
	{draw.Key1, draw.KeyNum1},
	{draw.Key2, draw.KeyNum2},
	{draw.Key3, draw.KeyNum3},
	{draw.Key4, draw.KeyNum4},
	{draw.Key5, draw.KeyNum5},
	{draw.Key6, draw.KeyNum6},
	{draw.Key7, draw.KeyNum7},
	{draw.Key8, draw.KeyNum8},
	{draw.Key9, draw.KeyNum9},
}

func (s *playingState) update(window draw.Window) state {
	// handle input
	if window.WasKeyPressed(draw.KeyEscape) {
		// TODO eventually go to a pause menu here
		window.Close()
	}
	keys := fireKeys[s.assignment.answer]
	if window.WasKeyPressed(keys[0]) || window.WasKeyPressed(keys[1]) {
		s.assignment = s.generator.generate(rand.Int)
	}
	if window.IsKeyDown(draw.KeyLeft) || window.IsKeyDown(draw.KeyA) {
		s.playerX -= playerSpeed
		s.playerFacingLeft = true
	}
	if window.IsKeyDown(draw.KeyRight) || window.IsKeyDown(draw.KeyD) {
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
