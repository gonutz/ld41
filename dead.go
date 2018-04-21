package main

import (
	"fmt"
	"github.com/gonutz/prototype/draw"
	"math/rand"
	"sort"
	"time"
)

type deadState struct {
	blink          int
	restartVisible bool
}

func (s *deadState) enter(oldState state) {
	s.restartVisible = true
	s.blink = 0
	if oldState == playing {
		score := playing.score
		highscores := loadHighScores()
		highscores = append(highscores, highscore{
			score: score,
			// TODO enter name here
			name: fmt.Sprintf("name %d", rand.Intn(10000)),
		})
		sort.Stable(byScore(highscores))
		const maxScores = 5
		if len(highscores) > maxScores {
			highscores = highscores[:maxScores]
		}
		saveHighScores(highscores)
	}
}

func (*deadState) leave() {}

func (s *deadState) update(window draw.Window) state {
	var nextState state = dead
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
