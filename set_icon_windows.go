package main

import "github.com/gonutz/w32"

func setIcon() {
	// the icon is contained in the .exe file as a resource, load it and set it
	// as the window icon so it appears in the top-left corner of the window and
	// when you alt+tab between windows
	const iconResourceID = 10
	iconHandle := w32.LoadImage(
		w32.GetModuleHandle(""),
		w32.MakeIntResource(iconResourceID),
		w32.IMAGE_ICON,
		0,
		0,
		w32.LR_DEFAULTSIZE|w32.LR_SHARED,
	)
	if iconHandle != 0 {
		window := w32.GetActiveWindow()
		w32.SendMessage(window, w32.WM_SETICON, w32.ICON_SMALL, uintptr(iconHandle))
		w32.SendMessage(window, w32.WM_SETICON, w32.ICON_SMALL2, uintptr(iconHandle))
		w32.SendMessage(window, w32.WM_SETICON, w32.ICON_BIG, uintptr(iconHandle))
	}
}
