//go:build js

package main

import "syscall/js"

func loadHighScores() []highscore {
	text := js.Global().Get("localStorage").Call("getItem", "brainless_jogging_highscores").String()
	return parseHighscores(text)
}

func saveHighScores(scores []highscore) {
	text := highscoresToString(scores)
	js.Global().Get("localStorage").Call("setItem", "brainless_jogging_highscores", text)
}
