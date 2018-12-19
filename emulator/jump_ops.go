package main

import "fmt"

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func jr(state *State) int {
	offset := int8(state.incrementPC())
	state.relativeJump(int(offset))

	if printInstructions {
		fmt.Printf("JR %#x\n", offset)
	}

	return 12
}

// jrIfFlag loads an offset value, then jumps to the operation at address PC +
// offset if the given flag is at the expected setting.
func jrIfFlag(state *State, flagMask uint8, isSet bool) int {
	flagState := state.regs8[regF].get()&flagMask == flagMask
	offset := int8(state.incrementPC())

	if printInstructions {
		conditional := getConditionalStr(flagMask, isSet)
		fmt.Printf("JR %v,%#x\n", conditional, offset)
	}

	if flagState == isSet {
		state.relativeJump(int(offset))
		return 12
	}
	// A relative jump didn't happen, so the instruction took fewer cycles
	return 8
}

// jp loads a 16-bit address and jumps to it.
func jp(state *State) int {
	address := combine16(state.incrementPC(), state.incrementPC())
	state.regs16[regPC].set(address)

	if printInstructions {
		fmt.Printf("JP %#x\n", address)
	}
	return 16
}

// jpIfFlag loads a 16-bit address and jumps to it if the given flag is at the
// expected setting.
func jpIfFlag(state *State, flagMask uint8, isSet bool) int {
	flagState := state.regs8[regF].get()&flagMask == flagMask
	address := combine16(state.incrementPC(), state.incrementPC())

	if printInstructions {
		conditional := getConditionalStr(flagMask, isSet)
		fmt.Printf("JP %v,%#x\n", conditional, address)
	}

	if flagState == isSet {
		state.regs16[regPC].set(address)
		return 16
	}
	// A jump didn't happen, so the instruction took fewer cycles
	return 12
}

// jpToHL jumps to the address specified by register HL.
func jpToHL(state *State) int {
	hlVal := state.regs16[regHL].get()

	state.regs16[regPC].set(hlVal)

	if printInstructions {
		fmt.Printf("JP (%v)\n", regHL)
	}
	return 4
}
