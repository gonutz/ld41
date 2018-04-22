package main

import "github.com/gonutz/prototype/draw"

type instructionsState struct{}

func (*instructionsState) enter(state) {}
func (*instructionsState) leave()      {}

func (*instructionsState) update(window draw.Window) state {
	if window.WasKeyPressed(draw.KeyEscape) {
		return menu
	}
	if window.WasKeyPressed(draw.KeyEnter) || window.WasKeyPressed(draw.KeyNumEnter) {
		return playing
	}
	const (
		text = `
   Solve math problems.
      Shoot zombies.
         Survive!

  Enter the solution to
the calculation above your
 head to shoot your rifle.

 Failing delays your next
          shot.

 Use the Left/Right arrow 
   keys or A/D to move.


   Press ENTER to play
`
		scale = 2
	)
	w, h := window.GetScaledTextSize(text, scale)
	window.DrawScaledText(text, (windowW-w)/2, (windowH-h)/2, scale, draw.White)

	return instructions
}
