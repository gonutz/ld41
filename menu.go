package main

import "github.com/gonutz/prototype/draw"

type menuState struct {
	hotItem int
	items   []string
}

func (s *menuState) enter(state) {
	s.hotItem = 0
	s.items = []string{
		"Start Game",
		"High Scores",
		"Quit",
	}
}

func (*menuState) leave() {}

func (s *menuState) update(window draw.Window) state {
	var nextState state = menu
	if window.WasKeyPressed(draw.KeyEscape) {
		window.Close()
	}
	oldItem := s.hotItem
	if window.WasKeyPressed(draw.KeyDown) {
		s.hotItem = (s.hotItem + 1) % len(s.items)
	}
	if window.WasKeyPressed(draw.KeyUp) {
		s.hotItem = (s.hotItem + len(s.items) - 1) % len(s.items)
	}
	if s.hotItem != oldItem {
		window.PlaySoundFile(file("menu beep.wav"))
	}
	if window.WasKeyPressed(draw.KeyEnter) || window.WasKeyPressed(draw.KeyNumEnter) {
		switch s.hotItem {
		case 0:
			nextState = playing
		case 1:
			nextState = dead
		case 2:
			window.Close()
		}
	}
	// render
	const textScale = 3
	for i, item := range s.items {
		w, h := window.GetScaledTextSize(item, textScale)
		x := (windowW - w) / 2
		y := (windowH-h*len(s.items))/2 + i*h
		if i == s.hotItem {
			window.FillRect(x-20, y, w+40, h, draw.DarkRed)
		}
		window.DrawScaledText(item, x, y, textScale, draw.White)
	}
	return nextState
}
