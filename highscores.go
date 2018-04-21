package main

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

type highscore struct {
	score int
	name  string
}

type byScore []highscore

func (x byScore) Len() int           { return len(x) }
func (x byScore) Less(i, j int) bool { return x[i].score > x[j].score }
func (x byScore) Swap(i, j int)      { x[i], x[j] = x[j], x[i] }

func loadHighScores() []highscore {
	path := filepath.Join(dataFolder, "ld41")
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil
	}
	s := string(data)
	lines := strings.Split(s, "\n")
	var scores []highscore
	for _, line := range lines {
		if line != "" {
			split := strings.Index(line, " ")
			if split == -1 {
				panic("invalid highscore file")
			}
			score, err := strconv.Atoi(line[:split])
			if err != nil {
				panic("invalid highscore file")
			}
			scores = append(scores, highscore{
				score: score,
				name:  line[split+1:],
			})
		}
	}
	return scores
}

func saveHighScores(scores []highscore) {
	path := filepath.Join(dataFolder, "ld41")
	var lines []string
	for _, s := range scores {
		lines = append(lines, fmt.Sprintf("%d %s", s.score, s.name))
	}
	ioutil.WriteFile(path, []byte(strings.Join(lines, "\n")), 0666)
}
