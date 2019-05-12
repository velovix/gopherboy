package gameboy

// makeRES creates an instruction that sets the specified bit of the given
// register to zero.
func makeRES(bitNum uint8, reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction and do operation

		reg.set(reg.get() & ^(0x1 << bitNum))

		return nil
	}
}

// makeRESMemHL creates an instruction that sets the specified bit of the value
// at the address specified by register HL to zero.
func makeRESMemHL(bitNum uint8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 2: Read from memory

			memVal := state.mmu.at(state.regHL.get())

			return func(state *State) instruction {
				// M-Cycle 3: Do operation and write to memory

				memVal &= ^(0x1 << bitNum)
				state.mmu.set(state.regHL.get(), memVal)

				return nil
			}
		}
	}
}

// makeBIT creates an instruction that checks the given bit of the given
// register value.
func makeBIT(bitNum uint8, reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction and do operation

		bitSet := reg.get()&(0x1<<bitNum) == (0x1 << bitNum)

		state.setZeroFlag(!bitSet)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)

		return nil
	}
}

// makeBITMemHL creates an instruction that checks the given bit of the value
// at the address specified by register HL.
func makeBITMemHL(bitNum uint8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 2: Read from memory and do operation

			memVal := state.mmu.at(state.regHL.get())

			bitSet := memVal&(0x1<<bitNum) == (0x1 << bitNum)

			state.setZeroFlag(!bitSet)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(true)

			return nil
		}
	}
}

// makeSWAP creates an instruction that swaps the upper and lower nibbles of
// the given register.
func makeSWAP(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		lower, upper := split(reg.get())
		reg.set(combine(upper, lower))

		state.setZeroFlag(reg.get() == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		return nil
	}
}

// swapMemHL swaps the upper and lower nibbles of the value in memory at the
// address specified by register HL.
func swapMemHL(state *State) instruction {
	// M-Cycle 1: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory

		memVal := state.mmu.at(state.regHL.get())

		return func(state *State) instruction {
			// M-Cycle 3: Do operation and write to memory

			lower, upper := split(memVal)
			memVal = combine(upper, lower)
			state.mmu.set(state.regHL.get(), memVal)

			state.setZeroFlag(memVal == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(false)
			state.setCarryFlag(false)

			return nil
		}
	}

}

// makeSET creates an instruction that sets the specified bit of the given
// register to one.
func makeSET(bitNum uint8, reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction and do operation

		reg.set(reg.get() | 0x1<<bitNum)

		return nil
	}
}

// makeSETMemHL creates an instruction that sets the specified bit of the value
// at the address specified by register HL to one.
func makeSETMemHL(bitNum uint8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 2: Read from memory

			memVal := state.mmu.at(state.regHL.get())

			return func(state *State) instruction {
				// M-Cycle 3: Write to memory

				memVal |= 0x1 << bitNum
				state.mmu.set(state.regHL.get(), memVal)

				return nil
			}
		}
	}
}
