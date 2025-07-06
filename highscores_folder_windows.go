package main

import (
	"os"
	"path/filepath"
)

func highscoresPath() string {
	return filepath.Join(os.Getenv("APPDATA"), "brainless_jogging_highscores")
}
