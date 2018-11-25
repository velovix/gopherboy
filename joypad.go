package main

import (
	"time"

	"github.com/veandco/go-sdl2/sdl"
)

// eventProcessDelay is the time in between keyboard event processing.
const eventProcessDelay = time.Second / 60

// joypad handles user input events and exposes them to the ROM.
type joypad struct {
	env *environment

	// The last known key states, where the map's key is a scan code and the
	// map's value is true if pressed.
	keyStates map[sdl.Scancode]bool

	// The last time keyboard events were processed.
	lastEventProcess time.Time
}

// newJoypad creates a new joypad manager object.
func newJoypad(env *environment) *joypad {
	j := &joypad{env: env}

	// Initialize all relevant keyboard values
	j.keyStates = map[sdl.Scancode]bool{
		sdl.SCANCODE_W:     false, // Start
		sdl.SCANCODE_Q:     false, // Select
		sdl.SCANCODE_Z:     false, // B
		sdl.SCANCODE_X:     false, // A
		sdl.SCANCODE_DOWN:  false, // Down
		sdl.SCANCODE_UP:    false, // Up
		sdl.SCANCODE_LEFT:  false, // Left
		sdl.SCANCODE_RIGHT: false, // Right
	}

	// Subscribe to P1 address writes
	j.env.mmu.subscribeTo(p1Addr, j.onP1Write)

	return j
}

// tick updates the P1 register value based on the current state of the user's
// input device and the game's configuration of this register. May also
// generate a P10-P13 interrupt if a button is pressed and this interrupt is
// enabled.
func (j *joypad) tick() {
	if time.Since(j.lastEventProcess) < eventProcessDelay {
		return
	}
	j.lastEventProcess = time.Now()

	buttonPressed := false

	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event := event.(type) {
		case *sdl.KeyboardEvent:
			scancode := event.Keysym.Scancode
			if _, ok := j.keyStates[scancode]; ok {
				if event.State == sdl.PRESSED {
					buttonPressed = true
					j.keyStates[scancode] = true
				} else {
					j.keyStates[scancode] = false
				}
			}
		}
	}

	if j.env.stopped && buttonPressed {
		// Exit stop mode on button press
		j.env.stopped = false
	}

	// Generate an interrupt if any new buttons have been pressed
	p10ThruP13InterruptEnabled := j.env.mmu.at(ieAddr)&0x10 == 0x10
	if buttonPressed && j.env.interruptsEnabled && p10ThruP13InterruptEnabled {
		j.env.mmu.set(ifAddr, j.env.mmu.at(ifAddr)|0x10)
	}
}

// onP1Write is called when the P1 register is written to. This triggers the
// joypad to update the memory register according to the requested data. This
// data is either the button states or the d-pad states.
func (j *joypad) onP1Write(addr uint16, writeVal uint8) uint8 {
	// Use bits 5 and 4 to decide what joypad input should be exposed. Note
	// that 0 means "select this" in this case.
	getDPadState := writeVal&0x10 == 0x00
	getButtonState := writeVal&0x20 == 0x00

	newP1 := writeVal

	// The unused first two bits of P1 are always high
	newP1 |= 0xC0

	// Set the state of the d-pad/buttons. Note again that 0 means the button
	// is pressed, not 1
	if getDPadState {
		// Check Down, Up, Left, and Right buttons
		newP1 |= buttonStatesToNibble(
			j.keyStates[sdl.SCANCODE_DOWN],
			j.keyStates[sdl.SCANCODE_UP],
			j.keyStates[sdl.SCANCODE_LEFT],
			j.keyStates[sdl.SCANCODE_RIGHT])
	} else if getButtonState {
		// Check Start, Select, B, and A buttons
		newP1 |= buttonStatesToNibble(
			j.keyStates[sdl.SCANCODE_W],
			j.keyStates[sdl.SCANCODE_Q],
			j.keyStates[sdl.SCANCODE_Z],
			j.keyStates[sdl.SCANCODE_X])
	} else {
		// No selection, provide nothing
		newP1 = 0
	}

	return newP1
}

// buttonStatesToNibble converts the given 4 button states into a nibble where
// a true value maps to a 0 and a false maps to 1.
func buttonStatesToNibble(states ...bool) uint8 {
	output := uint8(0x0)

	for i, state := range states {
		if !state {
			output |= 0x1 << uint(3-i)
		}
	}

	return output
}
