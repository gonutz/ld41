package main

import (
	"github.com/gonutz/prototype/draw"
	"math/rand"
)

const (
	playerSpeed        = 4
	playerW, playerH   = 158, 207
	bulletShootOffsetY = 103
	bulletW, bulletH   = 27, 9
)

type playingState struct {
	playerX, playerY int
	playerFacingLeft bool
	generator        mathGenerator
	assignment       assignment
	bullets          []bullet
}

func (s *playingState) enter(state) {
	s.playerX = windowW / 4
	s.playerY = 250
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
		s.shoot(window)
	}
	if window.IsKeyDown(draw.KeyLeft) || window.IsKeyDown(draw.KeyA) {
		s.playerX -= playerSpeed
		s.playerFacingLeft = true
	}
	if window.IsKeyDown(draw.KeyRight) || window.IsKeyDown(draw.KeyD) {
		s.playerX += playerSpeed
		s.playerFacingLeft = false
	}

	// update world
	n := 0
	for i := range s.bullets {
		b := &s.bullets[i]
		b.x += b.dx
		if (-100 <= b.x) && (b.x <= windowW+100) {
			s.bullets[n] = *b
			n++
		}
	}
	s.bullets = s.bullets[:n]

	// render
	// player
	window.FillRect(0, 0, windowW, windowH, draw.RGB(0, 0.5, 1))
	if s.playerFacingLeft {
		window.DrawImageFile(file("hero left.png"), s.playerX, s.playerY)
	} else {
		window.DrawImageFile(file("hero right.png"), s.playerX, s.playerY)
	}
	// zombie
	window.DrawImageFile(file("zombie left.png"), windowW-200, s.playerY)
	// bullets
	for _, b := range s.bullets {
		img := "bullet left.png"
		if b.dx > 0 {
			img = "bullet right.png"
		}
		window.DrawImageFile(file(img), b.x, b.y)
	}
	// assigment
	const mathScale = 2
	w, h := window.GetScaledTextSize(s.assignment.question, mathScale)
	window.DrawScaledText(
		s.assignment.question,
		s.playerX+(playerW-w)/2,
		s.playerY-h,
		mathScale,
		draw.White,
	)

	return playing
}

func (s *playingState) shoot(window draw.Window) {
	window.PlaySoundFile(file("shot.wav"))
	const bulletSpeed = 20
	var b bullet
	b.y = s.playerY + bulletShootOffsetY
	if s.playerFacingLeft {
		b.x = s.playerX
		b.dx = -bulletSpeed
	} else {
		b.x = s.playerX + playerW - bulletW
		b.dx = bulletSpeed
	}
	s.bullets = append(s.bullets, b)
	s.assignment = s.generator.generate(rand.Int)
}

type bullet struct {
	x, y int
	dx   int
}
