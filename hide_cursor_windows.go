package main

import "github.com/gonutz/w32"

func hideCursor() {
	w32.ShowCursor(false)
}

func showCursor() {
	w32.ShowCursor(true)
}
