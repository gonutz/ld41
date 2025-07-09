package main

import (
	"embed"
	"io"
	"math/rand"
	"time"

	"github.com/gonutz/prototype/draw"
)

//go:embed rsc/*
var rsc embed.FS

const (
	windowTitle      = "No-Brain Jogging"
	windowW, windowH = 1200, 600
	musicLength      = 8081 * time.Millisecond
)

type state interface {
	enter(from state)
	update(window draw.Window) state
	leave()
}

// all game states
var (
	menu         = &menuState{}
	playing      = &playingState{}
	dead         = &deadState{}
	instructions = &instructionsState{}
)

func main() {
	draw.OpenFile = func(path string) (io.ReadCloser, error) {
		return rsc.Open("rsc/" + path)
	}

	rand.Seed(time.Now().UnixNano())

	var state state = menu
	state.enter(nil)

	var musicStart time.Time
	firstFrame := true

	check(draw.RunWindow(windowTitle, windowW, windowH, func(window draw.Window) {
		if firstFrame {
			window.SetIcon("icon.png")
			window.ShowCursor(false)
			preloadAssets(window)
			firstFrame = false
		}

		newState := state.update(window)
		if state != newState {
			state.leave()
			newState.enter(state)
		}
		state = newState

		now := time.Now()
		if now.Sub(musicStart) >= musicLength {
			window.PlaySoundFile("music.wav")
			musicStart = now
		}
	}))
}

func preloadAssets(window draw.Window) {
	files, _ := rsc.ReadDir("rsc")
	for _, file := range files {
		window.DrawImageFile(file.Name(), -9999, -9999)
	}
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
