//go:build !windows && !js

package main

import (
	"os"
	"path/filepath"
)

func highscoresPath() string {
	dir := "."

	if exe, err := os.Executable(); err == nil {
		dir = filepath.Dir(exe)
	}

	return filepath.Join(dir, "brainless_jogging_highscores")
}
