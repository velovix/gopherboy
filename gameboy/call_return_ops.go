package gameboy

import "fmt"

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address.
func call(state *State) int {
	address := combine16(state.incrementPC(), state.incrementPC())
	state.pushToStack16(state.regs16[regPC].get())

	state.regs16[regPC].set(address)

	if printInstructions {
		fmt.Printf("CALL %#x\n", address)
	}
	return 24
}

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address if the given flag is at the
// expected setting.
func callIfFlag(state *State, flagMask uint8, isSet bool) int {
	flagState := state.regs8[regF].get()&flagMask == flagMask
	address := combine16(state.incrementPC(), state.incrementPC())

	if printInstructions {
		conditional := getConditionalStr(flagMask, isSet)
		fmt.Printf("CALL %v,%#x\n", conditional, address)
	}

	if flagState == isSet {
		state.pushToStack16(state.regs16[regPC].get())
		state.regs16[regPC].set(address)

		return 24
	}
	// A call didn't happen, so the instruction took fewer cycles
	return 12
}

// ret pops a 16-bit address from the stack and jumps to it.
func ret(state *State) int {
	addr := state.popFromStack16()
	state.regs16[regPC].set(addr)

	if printInstructions {
		fmt.Printf("RET\n")
	}
	return 16
}

// retIfFlag pops a 16-bit address from the stack and jumps to it, but only if
// the given flag is at the expected value.
func retIfFlag(state *State, flagMask uint8, isSet bool) int {
	flagState := state.regs8[regF].get()&flagMask == flagMask

	var opClocks int
	if flagState == isSet {
		addr := state.popFromStack16()
		state.regs16[regPC].set(addr)
		opClocks = 20
	} else {
		opClocks = 8
	}

	if printInstructions {
		conditional := getConditionalStr(flagMask, isSet)
		fmt.Printf("RET %v\n", conditional)
	}
	return opClocks
}

// reti pops a 16-bit address from the stack and jumps to it, then enables
// interrupts.
func reti(state *State) int {
	addr := state.popFromStack16()
	state.regs16[regPC].set(addr)

	state.interruptsEnabled = true

	if printInstructions {
		fmt.Printf("RETI\n")
	}
	return 16
}

// rst pushes the current program counter to the stack and jumps to the given
// address.
func rst(state *State, address uint16) int {
	state.pushToStack16(state.regs16[regPC].get())

	state.regs16[regPC].set(address)

	if printInstructions {
		fmt.Printf("RST %#x\n", address)
	}
	return 16
}