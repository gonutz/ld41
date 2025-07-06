package main

import (
	"fmt"
	"strconv"
	"strings"
)

type highscore struct {
	score int
	name  string
	id    int // id is used only temporarily in the code, do not save/load it
}

type byScore []highscore

func (x byScore) Len() int           { return len(x) }
func (x byScore) Less(i, j int) bool { return x[i].score > x[j].score }
func (x byScore) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func parseHighscores(text string) []highscore {
	lines := strings.Split(text, "\n")
	var scores []highscore
	for _, line := range lines {
		cols := strings.SplitN(line, " ", 2)
		if len(cols) == 2 {
			score, _ := strconv.Atoi(cols[0])
			name := cols[1]
			if score > 0 {
				scores = append(scores, highscore{
					score: score,
					name:  name,
				})
			}
		}
	}
	return scores
}

func highscoresToString(scores []highscore) string {
	text := ""
	for _, s := range scores {
		text += fmt.Sprintf("%d %s\n", s.score, s.name)
	}
	return text
}
