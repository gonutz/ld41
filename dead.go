package main

import (
	"github.com/gonutz/prototype/draw"
	"time"
)

type deadState struct {
	blink          int
	restartVisible bool
}

func (s *deadState) enter(state) {
	s.restartVisible = true
	s.blink = 0
}

func (*deadState) leave() {}

func (s *deadState) update(window draw.Window) state {
	nextState := dead
	if window.WasKeyPressed(draw.KeyEnter) || window.WasKeyPressed(draw.KeyNumEnter) {
		nextState = playing
	}
	if window.WasKeyPressed(draw.KeyEscape) {
		window.Close()
	}

	s.blink--
	if s.blink <= 0 {
		s.restartVisible = !s.restartVisible
		if s.restartVisible {
			s.blink = frames(700 * time.Millisecond)
		} else {
			s.blink = frames(400 * time.Millisecond)
		}
	}

	const (
		title     = "You were eaten alive!"
		msg       = "Press ENTER to restart"
		textScale = 3
	)
	w, h := window.GetScaledTextSize(title, textScale)
	window.DrawScaledText(title, (windowW-w)/2, windowH/2-h, textScale, draw.White)
	if s.restartVisible {
		w, h := window.GetScaledTextSize(msg, textScale)
		window.DrawScaledText(msg, (windowW-w)/2, windowH/2+h, textScale, draw.White)
	}
	return nextState
}
