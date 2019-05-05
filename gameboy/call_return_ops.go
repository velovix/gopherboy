package gameboy

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address.
func call(state *State) int {
	address := combine16(state.incrementPC(), state.incrementPC())
	state.pushToStack16(state.regPC.get())

	state.regPC.set(address)

	return 24
}

// makeCALLIfFlag creates an instruction that loads a 16-bit address, pushes
// the address of the next instruction onto the stack, and jumps to the loaded
// address if the given flag is at the expected setting.
func makeCALLIfFlag(flagMask uint8, isSet bool) instruction {
	return adapter(func(state *State) int {
		flagState := state.regF.get()&flagMask == flagMask
		address := combine16(state.incrementPC(), state.incrementPC())

		if flagState == isSet {
			state.pushToStack16(state.regPC.get())
			state.regPC.set(address)

			return 24
		}
		// A call didn't happen, so the instruction took fewer cycles
		return 12
	})
}

// ret pops a 16-bit address from the stack and jumps to it.
func ret(state *State) int {
	addr := state.popFromStack16()
	state.regPC.set(addr)

	return 16
}

// makeRETIfFlag creates an instruction that pops a 16-bit address from the
// stack and jumps to it, but only if the given flag is at the expected value.
func makeRETIfFlag(flagMask uint8, isSet bool) instruction {
	return adapter(func(state *State) int {
		flagState := state.regF.get()&flagMask == flagMask

		var opClocks int
		if flagState == isSet {
			addr := state.popFromStack16()
			state.regPC.set(addr)
			opClocks = 20
		} else {
			opClocks = 8
		}

		return opClocks
	})
}

// reti pops a 16-bit address from the stack and jumps to it, then enables
// interrupts.
func reti(state *State) int {
	addr := state.popFromStack16()
	state.regPC.set(addr)

	state.interruptsEnabled = true

	return 16
}

// makeRST creates an instruction that pushes the current program counter to
// the stack and jumps to the given address.
func makeRST(address uint16) instruction {
	return adapter(func(state *State) int {
		state.pushToStack16(state.regPC.get())

		state.regPC.set(address)

		return 16
	})
}
