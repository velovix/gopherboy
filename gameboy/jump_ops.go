package gameboy

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func jr(state *State) int {
	// M-Cycle 0: This instruction doesn't do anything

	// M-Cycle 1: Load the offset and apply it to the PC
	/*return func(state *State) instruction {
		offset := int8(state.incrementPC())
		relativeJump(state, offset)

		// M-Cycle 2: Internal delay
		return func(state *State) instruction {
			return nil
		}
	}*/

	offset := int8(state.incrementPC())
	relativeJump(state, offset)
	return 12
}

// makeJRIfFlag creates an instruction that loads an offset value, then jumps
// to the operation at address PC + offset if the given flag is at the expected
// setting.
func makeJRIfFlag(flagMask uint8, isSet bool) instruction {
	return adapter(func(state *State) int {
		flagState := state.regF.get()&flagMask == flagMask
		offset := int8(state.incrementPC())

		if flagState == isSet {
			relativeJump(state, offset)
			return 12
		}
		// A relative jump didn't happen, so the instruction took fewer cycles
		return 8
	})
}

// jp loads a 16-bit address and jumps to it.
func jp(state *State) instruction {
	// M-Cycle 0: This instruction doesn't do anything

	// M-Cycle 1: Load the least significant bit of the target address
	return func(state *State) instruction {
		lsb := state.incrementPC()

		// M-Cycle 2: Load the most significant bit of the target address
		return func(state *State) instruction {
			msb := state.incrementPC()
			state.regPC.set(combine16(lsb, msb))

			// M-Cycle 3: Internal delay
			return func(state *State) instruction {
				return nil
			}
		}
	}
}

// makeJPIfFlag creates an instruction that loads a 16-bit address and jumps to
// it if the given flag is at the expected setting.
func makeJPIfFlag(flagMask uint8, isSet bool) instruction {
	return adapter(func(state *State) int {
		flagState := state.regF.get()&flagMask == flagMask
		address := combine16(state.incrementPC(), state.incrementPC())

		if flagState == isSet {
			state.regPC.set(address)
			return 16
		}
		// A jump didn't happen, so the instruction took fewer cycles
		return 12
	})
}

// jpToHL jumps to the address specified by register HL.
func jpToHL(state *State) instruction {
	hlVal := state.regHL.get()

	state.regPC.set(hlVal)

	return nil
}

// relativeJump moves the program counter by the given signed value.
func relativeJump(state *State, offset int8) {
	pc := uint16(int(state.regPC.get()) + int(offset))
	state.regPC.set(pc)
}
