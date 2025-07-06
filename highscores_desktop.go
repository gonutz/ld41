//go:build !js

package main

import "os"

func loadHighScores() []highscore {
	data, err := os.ReadFile(highscoresPath())
	if err != nil {
		return nil
	}
	return parseHighscores(string(data))
}

func saveHighScores(scores []highscore) {
	os.WriteFile(highscoresPath(), []byte(highscoresToString(scores)), 0666)
}
