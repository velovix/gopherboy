package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"github.com/velovix/gopherboy/gameboy"
)

var scancodeToButton = map[sdl.Scancode]gameboy.Button{
	sdl.SCANCODE_W:     gameboy.ButtonStart,
	sdl.SCANCODE_Q:     gameboy.ButtonSelect,
	sdl.SCANCODE_Z:     gameboy.ButtonB,
	sdl.SCANCODE_X:     gameboy.ButtonA,
	sdl.SCANCODE_DOWN:  gameboy.ButtonDown,
	sdl.SCANCODE_UP:    gameboy.ButtonUp,
	sdl.SCANCODE_LEFT:  gameboy.ButtonLeft,
	sdl.SCANCODE_RIGHT: gameboy.ButtonRight,
}

type inputDriver struct {
	buttonStates map[gameboy.Button]bool
}

func newInputDriver() *inputDriver {
	var driver inputDriver

	// Initialize all relevant joypad values
	driver.buttonStates = map[gameboy.Button]bool{
		gameboy.ButtonStart:  false,
		gameboy.ButtonSelect: false,
		gameboy.ButtonB:      false,
		gameboy.ButtonA:      false,
		gameboy.ButtonDown:   false,
		gameboy.ButtonUp:     false,
		gameboy.ButtonLeft:   false,
		gameboy.ButtonRight:  false,
	}

	return &driver
}

func (driver *inputDriver) State(btn gameboy.Button) bool {
	return driver.buttonStates[btn]
}

// Update polls the keyboard for the latest input events and updates the button
// states accordingly. Returns true if a new button was pressed since the last
// call to update.
func (driver *inputDriver) Update() bool {
	buttonPressed := false

	doOnMainThread(func() {
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event := event.(type) {
			case *sdl.KeyboardEvent:
				btn := scancodeToButton[event.Keysym.Scancode]

				if _, ok := driver.buttonStates[btn]; ok {
					if event.State == sdl.PRESSED {
						buttonPressed = true
						driver.buttonStates[btn] = true
					} else {
						driver.buttonStates[btn] = false
					}
				}
			}
		}
	}, false)

	return buttonPressed
}
