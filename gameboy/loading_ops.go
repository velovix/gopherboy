package gameboy

// makeLD creates an instruction that loads the value of reg2 into reg1.
func makeLD(reg1, reg2 register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Do operation

		reg1.set(reg2.get())

		return nil
	}
}

// ldHLToSP puts the value of register HL into register SP.
func ldHLToSP(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		hlVal := state.regHL.get()

		state.regSP.set(hlVal)

		return nil
	}
}

// makeLDToMem creates an instruction that loads the value of reg2 into the
// memory address specified by reg1.
func makeLDToMem(reg1 register16, reg2 register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Write to memory

			state.mmu.set(reg1.get(), reg2.get())

			return nil
		}
	}
}

// makeLDFromMem creates an instruction that loads the value in the memory
// address specified by reg2 into reg1.
func makeLDFromMem(reg1 register8, reg2 register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Read from memory into register

			val := state.mmu.at(reg2.get())
			reg1.set(val)

			return nil
		}
	}
}

// makeLD8BitImm creates an instruction that loads an 8-bit immediate value
// into the given register.
func makeLD8BitImm(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Read from PC position and set register

			imm := state.incrementPC()

			reg.set(imm)

			return nil
		}
	}
}

// makeLD16BitImm creates an instruction that loads a 16-bit immediate value
// into the specified 16-bit register.
func makeLD16BitImm(reg register16) func(*State) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Read least significant byte from PC position

			immLower := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 2: Read most significant byte from PC position and
				//            set the register

				immUpper := state.incrementPC()

				imm := combine16(immLower, immUpper)
				reg.set(imm)

				return nil
			}
		}
	}
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read least significant byte from PC position

		immLower := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Read most significant byte from PC position

			immUpper := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 3: Write register to memory

				imm := combine16(immLower, immUpper)
				aVal := state.regA.get()

				state.mmu.set(imm, aVal)

				return nil
			}
		}
	}
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read least significant byte from PC position

		immLower := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Read most significant byte from PC position

			immUpper := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 3: Write register to memory

				imm := combine16(immLower, immUpper)
				memVal := state.mmu.at(imm)

				state.regA.set(memVal)

				return nil
			}
		}
	}
}

// ld8BitImmToMemHL loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value

		imm := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Write to memory

			state.mmu.set(state.regHL.get(), imm)

			return nil
		}
	}
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read least significant byte at PC position

		immLower := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Read most significant byte at PC position

			immUpper := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 3: Save lower byte of stack pointer to memory

				imm := combine16(immLower, immUpper)

				lowerSP, _ := split16(state.regSP.get())
				state.mmu.set(imm, lowerSP)

				return func(state *State) instruction {
					// M-Cycle 4: Save upper byte of stack pointer to memory

					_, upperSP := split16(state.regSP.get())
					state.mmu.set(imm+1, upperSP)

					return nil
				}
			}
		}
	}
}

// ldToMemC saves the value of register A at the memory address
// 0xFF00+register C.
func ldToMemC(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Write to memory

		aVal := state.regA.get()
		addr := uint16(state.regC.get()) + 0xFF00

		state.mmu.set(addr, aVal)

		return nil
	}
}

// ldFromMemC loads the value at memory address 0xFF00 + register C into
// register A.
func ldFromMemC(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory

		addr := uint16(state.regC.get()) + 0xFF00
		memVal := state.mmu.at(addr)

		state.regA.set(memVal)

		return nil
	}
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Write to memory

		state.mmu.set(state.regHL.get(), state.regA.get())

		state.regHL.set(state.regHL.get() + 1)

		return nil
	}
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Write to memory

		state.mmu.set(state.regHL.get(), state.regA.get())

		state.regHL.set(state.regHL.get() - 1)

		return nil
	}
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory

		memVal := state.mmu.at(state.regHL.get())
		state.regA.set(memVal)

		state.regHL.set(state.regHL.get() + 1)

		return nil
	}
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory

		memVal := state.mmu.at(state.regHL.get())
		state.regA.set(memVal)

		state.regHL.set(state.regHL.get() - 1)

		return nil
	}
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read offset from PC location

		offset := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Write to memory

			state.mmu.set(0xFF00+uint16(offset), state.regA.get())

			return nil
		}
	}
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read offset from PC location

		offset := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Read from memory

			fromMem := state.mmu.at(0xFF00 + uint16(offset))
			state.regA.set(fromMem)

			return nil
		}
	}
}

// ldhl loads an 8-bit immediate value and adds it to the stack pointer. The
// result is stored in register HL.
func ldhl(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation

		immUnsigned := state.incrementPC()
		imm := int8(immUnsigned)

		spVal := state.regSP.get()

		// This instruction's behavior for the carry and half carry flags is very
		// weird.
		//
		// When checking for a carry and half carry, the immediate value is treated
		// as _unsigned_ for some reason and only the lowest 8 bits of the stack
		// pointer are considered. HL does not play into this calculation at all.
		lowerSP, _ := split16(spVal)
		state.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
		state.setCarryFlag(isCarry(lowerSP, immUnsigned))

		state.regHL.set(uint16(int(imm) + int(spVal)))

		state.setZeroFlag(false)
		state.setSubtractFlag(false)

		return func(state *State) instruction {
			// M-Cycle 2: Internal delay

			return nil
		}
	}
}

// makePUSH creates an instruction that decrements the stack pointer by 2, then
// puts the value of the given register at its position.
func makePUSH(reg register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch operation

		return func(state *State) instruction {
			// M-Cycle 1: Internal delay

			return func(state *State) instruction {
				// M-Cycle 2: Write upper byte to stack

				_, upper := split16(reg.get())
				state.pushToStack(upper)

				return func(state *State) instruction {
					// M-Cycle 3: Write lower byte to stack

					lower, _ := split16(reg.get())
					state.pushToStack(lower)

					return nil
				}
			}
		}
	}
}

// makePOP creates an instruction that loads the two bytes at the top of the
// stack in the given register and increments the stack pointer by 2.
func makePOP(reg register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Pop lower byte from stack

			reg.setLower(state.popFromStack())

			return func(state *State) instruction {
				// M-Cycle 2: Pop upper byte from stack

				reg.setUpper(state.popFromStack())

				return nil
			}
		}
	}
}
