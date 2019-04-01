package gameboy

import "fmt"

// nop does nothing.
func nop(state *State) int {
	return 4
}

// di sets the master interrupt flag to false, disabling all interrupt
// handling.
func di(state *State) int {
	state.interruptsEnabled = false
	// Cancel a delayed interrupt enable request if any
	state.enableInterruptsTimer = 0

	return 4
}

// ei sets the master interrupt flag to true, but not until the instruction
// after EI has been executed. Interrupts may still be disabled using the
// interrupt flags memory register, however.
func ei(state *State) int {
	state.enableInterruptsTimer = 2

	return 4
}

// halt stops running instructions until an interrupt is triggered.
func halt(state *State) int {
	state.halted = true

	return 4
}

// cpl inverts the value of register A.
func cpl(state *State) int {
	invertedA := ^state.regA.get()
	state.regA.set(invertedA)

	state.setHalfCarryFlag(true)
	state.setSubtractFlag(true)

	return 4
}

// ccf flips the carry flag.
func ccf(state *State) int {
	state.setCarryFlag(!state.getCarryFlag())

	state.setHalfCarryFlag(false)
	state.setSubtractFlag(false)

	return 4
}

// scf sets the carry flag to true.
func scf(state *State) int {
	state.setCarryFlag(true)
	state.setHalfCarryFlag(false)
	state.setSubtractFlag(false)

	return 4
}

// stop puts the Game Boy in stop mode. In this mode, the screen is blank and
// the CPU stops. Stop mode is exited when a button is pressed.
func stop(state *State) int {
	// For whatever reason, this instruction is two bytes in length
	state.incrementPC()

	fmt.Println("Switch to STOP mode")

	state.stopped = true

	return 4
}
