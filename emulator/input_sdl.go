// +build linux darwin windows

package main

import "github.com/veandco/go-sdl2/sdl"

type button sdl.Scancode

const (
	buttonStart  button = sdl.SCANCODE_W
	buttonSelect        = sdl.SCANCODE_Q
	buttonB             = sdl.SCANCODE_Z
	buttonA             = sdl.SCANCODE_X
	buttonDown          = sdl.SCANCODE_DOWN
	buttonUp            = sdl.SCANCODE_UP
	buttonLeft          = sdl.SCANCODE_LEFT
	buttonRight         = sdl.SCANCODE_RIGHT
)

type inputDriver struct {
	buttonStates map[button]bool
}

func newInputDriver() *inputDriver {
	var driver inputDriver

	// Initialize all relevant joypad values
	driver.buttonStates = map[button]bool{
		buttonStart:  false, // Start
		buttonSelect: false, // Select
		buttonB:      false, // B
		buttonA:      false, // A
		buttonDown:   false, // Down
		buttonUp:     false, // Up
		buttonLeft:   false, // Left
		buttonRight:  false, // Right
	}

	return &driver
}

// update polls the keyboard for the latest input events and updates the button
// states accordingly. Returns true if a new button was pressed since the last
// call to update.
func (driver *inputDriver) update() bool {
	buttonPressed := false

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event := event.(type) {
		case *sdl.KeyboardEvent:
			btn := button(event.Keysym.Scancode)

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

	return buttonPressed
}
