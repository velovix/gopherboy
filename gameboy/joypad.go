package gameboy

// eventProcessPeriod is the delay between input processes. This doesn't have
// any meaning to the Game Boy hardware, but limiting the input processing
// provides better performance.
const eventProcessPeriod = cpuClockRate / 60

// joypad handles user input events and exposes them to the ROM.
type joypad struct {
	state *State

	driver InputDriver

	// The number of cycles since the last event processing.
	lastEventProcess int
}

// newJoypad creates a new joypad manager object.
func newJoypad(state *State, driver InputDriver) *joypad {
	j := &joypad{
		state:  state,
		driver: driver,
	}

	// Subscribe to P1 address writes
	j.state.mmu.subscribeTo(p1Addr, j.onP1Write)

	return j
}

// tick updates the P1 register value based on the current state of the user's
// input device and the game's configuration of this register. May also
// generate a P10-P13 interrupt if a button is pressed and this interrupt is
// enabled.
func (j *joypad) tick(amount int) {
	j.lastEventProcess += amount
	if j.lastEventProcess < eventProcessPeriod {
		return
	}

	j.lastEventProcess -= eventProcessPeriod

	buttonPressed := j.driver.Update()

	if j.state.stopped && buttonPressed {
		// Exit stop mode on button press
		j.state.stopped = false
	}

	// Generate an interrupt if any new buttons have been pressed
	p10ThruP13InterruptEnabled := j.state.mmu.memory[ieAddr]&0x10 == 0x10
	if buttonPressed && j.state.interruptsEnabled && p10ThruP13InterruptEnabled {
		j.state.mmu.setNoNotify(ifAddr, j.state.mmu.at(ifAddr)|0x10)
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
			j.driver.State(ButtonDown),
			j.driver.State(ButtonUp),
			j.driver.State(ButtonLeft),
			j.driver.State(ButtonRight))
	} else if getButtonState {
		// Check Start, Select, B, and A buttons
		newP1 |= buttonStatesToNibble(
			j.driver.State(ButtonStart),
			j.driver.State(ButtonSelect),
			j.driver.State(ButtonB),
			j.driver.State(ButtonA))
	}
	// Otherwise, provide nothing

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
