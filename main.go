package main

import "github.com/gonutz/prototype/draw"

const (
	title            = "LD41 - TODO: name this game"
	windowW, windowH = 1000, 600
)

func main() {
	check(draw.RunWindow(title, windowW, windowH, func(window draw.Window) {
		if window.WasKeyPressed(draw.KeyEscape) {
			window.Close()
		}
	}))
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
