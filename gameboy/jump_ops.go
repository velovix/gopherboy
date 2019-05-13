package gameboy

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func jr(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value

		offset := int8(state.incrementPC())

		return func(state *State) instruction {
			// M-Cycle 2: Do operation

			relativeJump(state, offset)

			return nil
		}
	}
}

// makeJRIfFlag creates an instruction that loads an offset value, then jumps
// to the operation at address PC + offset if the given flag is at the expected
// setting.
func makeJRIfFlag(flagMask uint8, isSet bool) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and check condition
		offset := int8(state.incrementPC())

		flagState := state.regF.get()&flagMask == flagMask

		if flagState != isSet {
			// Condition evaluated to false, don't jump
			return nil
		}

		return func(state *State) instruction {
			// M-Cycle 2: Do operation

			relativeJump(state, offset)

			return nil
		}
	}
}

// jp loads a 16-bit address and jumps to it.
func jp(state *State) instruction {
	// M-Cycle 0: This instruction doesn't do anything

	return func(state *State) instruction {
		// M-Cycle 1: Load the least significant bit of the target address

		addressLower := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Load the most significant bit of the target address
			//            and set the program counter

			addressUpper := state.incrementPC()

			state.regPC.setLower(addressLower)
			state.regPC.setUpper(addressUpper)

			return func(state *State) instruction {
				// M-Cycle 3: Internal delay

				return nil
			}
		}
	}
}

// makeJPIfFlag creates an instruction that loads a 16-bit address and jumps to
// it if the given flag is at the expected setting.
func makeJPIfFlag(flagMask uint8, isSet bool) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Read lower byte of target address

			addressLower := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 2: Read upper byte of target address and check
				//            condition

				addressUpper := state.incrementPC()

				flagState := state.regF.get()&flagMask == flagMask
				if flagState != isSet {
					// Condition evaluated to false, don't jump
					return nil
				}

				return func(state *State) instruction {
					// M-Cycle 3: Set the PC

					state.regPC.setLower(addressLower)
					state.regPC.setUpper(addressUpper)

					return nil
				}
			}
		}
	}
}

// jpToHL jumps to the address specified by register HL.
func jpToHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	hlVal := state.regHL.get()

	state.regPC.set(hlVal)

	return nil
}

// relativeJump moves the program counter by the given signed value.
func relativeJump(state *State, offset int8) {
	pc := uint16(int(state.regPC.get()) + int(offset))
	state.regPC.set(pc)
}
