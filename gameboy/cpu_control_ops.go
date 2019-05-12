package gameboy

import "fmt"

// nop does nothing.
func nop(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return nil
}

// di sets the master interrupt flag to false, disabling all interrupt
// handling.
func di(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	state.interruptsEnabled = false
	// Cancel a delayed interrupt enable request if any
	state.enableInterruptsTimer = 0

	return nil
}

// ei sets the master interrupt flag to true, but not until the instruction
// after EI has been executed. Interrupts may still be disabled using the
// interrupt flags memory register, however.
func ei(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	state.enableInterruptsTimer = 2

	return nil
}

// halt stops running instructions until an interrupt is triggered.
func halt(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	state.halted = true

	return nil
}

// cpl inverts the value of register A.
func cpl(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	invertedA := ^state.regA.get()
	state.regA.set(invertedA)

	state.setHalfCarryFlag(true)
	state.setSubtractFlag(true)

	return nil
}

// ccf flips the carry flag.
func ccf(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	state.setCarryFlag(!state.getCarryFlag())

	state.setHalfCarryFlag(false)
	state.setSubtractFlag(false)

	return nil
}

// scf sets the carry flag to true.
func scf(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	state.setCarryFlag(true)
	state.setHalfCarryFlag(false)
	state.setSubtractFlag(false)

	return nil
}

// stop puts the Game Boy in stop mode. In this mode, the screen is blank and
// the CPU stops. Stop mode is exited when a button is pressed.
func stop(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	// For whatever reason, this instruction is two bytes in length
	state.incrementPC()

	fmt.Println("Switch to STOP mode")

	state.stopped = true

	return nil
}
