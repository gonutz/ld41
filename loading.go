package main

import (
	"github.com/gonutz/blob"
	"github.com/gonutz/payload"
	"github.com/gonutz/prototype/draw"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

type loadingState struct {
	assetsLoaded bool
}

func (s *loadingState) enter(state) {
	load, err := payload.Open()
	if err == nil {
		// this executable has a blob of assets attached, write them to disk for
		// the prototype library to use them
		data, err := blob.Open(load)
		check(err)
		dir, err := ioutil.TempDir("", "ld41_")
		check(err)
		cleanUpAssets = func() {
			os.RemoveAll(dir)
		}
		file = func(filename string) string {
			return filepath.Join(dir, filename)
		}
		go func() {
			defer func() { s.assetsLoaded = true }()
			for i := 0; i < data.ItemCount(); i++ {
				id := data.GetIDAtIndex(i)
				r, _ := data.GetByIndex(i)
				func() {
					f, err := os.Create(file(id))
					check(err)
					defer f.Close()
					_, err = io.Copy(f, r)
					check(err)
				}()
			}
		}()
	} else {
		// no payload in this executable, load from files
		file = func(f string) string { return filepath.Join("rsc", f) }
		s.assetsLoaded = true
	}
}

func (*loadingState) leave() {}

func (s *loadingState) update(window draw.Window) state {
	const (
		text      = "Loading..."
		textScale = 2
	)
	w, h := window.GetScaledTextSize(text, textScale)
	window.DrawScaledText(text, (windowW-w)/2, (windowH-h)/2, textScale, draw.White)
	if s.assetsLoaded {
		return playing
	} else {
		return loading
	}
}
