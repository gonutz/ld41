package main

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gonutz/prototype/draw"
)

const (
	maxHighScores   = 5
	maxNameLen      = 20
	cursorBlinkTime = 300 * time.Millisecond
)

type deadState struct {
	caption        string
	blink          int
	restartVisible bool
	highscores     []highscore
	editing        int
	cursorBlink    int
	cursorVisible  bool
	score          int
}

func (s *deadState) enter(oldState state) {
	s.score = -1
	s.restartVisible = true
	s.blink = 0
	s.editing = -1
	s.highscores = loadHighScores()
	if len(s.highscores) < maxHighScores {
		s.highscores = append(s.highscores, make([]highscore, maxHighScores-len(s.highscores))...)
	}
	s.caption = "High Scores"
	if oldState == playing {
		s.caption = "You were eaten alive!"
		score := playing.score
		s.score = score
		s.highscores = append(s.highscores, highscore{
			score: score,
			name:  "",
			id:    1,
		})
		sort.Stable(byScore(s.highscores))
		if len(s.highscores) > maxHighScores {
			s.highscores = s.highscores[:maxHighScores]
		}
		saveHighScores(s.highscores)
		s.editing = -1
		for i := range s.highscores {
			if s.highscores[i].id == 1 {
				s.editing = i
			}
		}
	}
	s.restartVisible = false
	s.cursorBlink = 0
	s.cursorVisible = false
}

func (*deadState) leave() {}

func (s *deadState) update(window draw.Window) state {
	var nextState state = dead
	// handle input
	if window.WasKeyPressed(draw.KeyEscape) {
		nextState = menu
	}
	if window.WasKeyPressed(draw.KeyEnter) || window.WasKeyPressed(draw.KeyNumEnter) {
		if s.editing != -1 {
			s.editing = -1
			saveHighScores(s.highscores)
			s.restartVisible = false
			s.blink = 0
		} else {
			nextState = playing
		}
	}
	// text input if editing high score name
	if s.editing != -1 {
		score := &s.highscores[s.editing]
		typed := window.Characters()
		for _, r := range typed {
			if len(score.name) < maxNameLen && (32 <= r) && (r <= 126) {
				score.name += string(r)
			}
			s.cursorVisible = true
			s.cursorBlink = frames(cursorBlinkTime)
		}
		if window.WasKeyPressed(draw.KeyBackspace) && score.name != "" {
			_, size := utf8.DecodeLastRuneInString(score.name)
			score.name = score.name[:len(score.name)-size]
			s.cursorVisible = true
			s.cursorBlink = frames(cursorBlinkTime)
		}
	}
	// update animations
	s.blink--
	if s.blink <= 0 {
		s.restartVisible = !s.restartVisible
		if s.restartVisible {
			s.blink = frames(700 * time.Millisecond)
		} else {
			s.blink = frames(400 * time.Millisecond)
		}
	}
	s.cursorBlink--
	if s.cursorBlink < 0 {
		s.cursorVisible = !s.cursorVisible
		s.cursorBlink = frames(cursorBlinkTime)
	}
	// render
	// highscores
	const scoreScale = 2
	lineW, lineH := window.GetScaledTextSize(
		strings.Repeat("A", maxNameLen+len("1.  25")),
		scoreScale,
	)
	scoresY := (windowH - 5*lineH) / 2
	for i, score := range s.highscores {
		name := score.name
		if i == s.editing && s.cursorVisible {
			name += "|"
		}
		if len(name) < maxNameLen {
			name += strings.Repeat(".", maxNameLen-len(name))
		}
		space := " "
		if len(name) > maxNameLen {
			space = ""
		}
		scoreText := fmt.Sprintf("%d. %s%s%d", i+1, name, space, score.score)
		window.DrawScaledText(scoreText, (windowW-lineW)/2, scoresY+i*lineH, scoreScale, draw.White)
	}
	// title and instructions
	const (
		msg       = "Press ENTER to play"
		textScale = 3
	)
	w, h := window.GetScaledTextSize(s.caption, textScale)
	window.DrawScaledText(s.caption, (windowW-w)/2, scoresY-h-50, textScale, draw.White)
	if s.editing == -1 && s.restartVisible {
		w, _ := window.GetScaledTextSize(msg, textScale)
		window.DrawScaledText(msg, (windowW-w)/2, scoresY+5*lineH+50, textScale, draw.White)
	}
	// score
	if s.score >= 0 {
		suffix := "s"
		if s.score == 1 {
			suffix = ""
		}
		text := fmt.Sprintf("You killed %d zombie%s", s.score, suffix)
		w, _ := window.GetScaledTextSize(text, textScale)
		window.DrawScaledText(text, (windowW-w)/2, 30, textScale, draw.DarkRed)
	}
	return nextState
}
