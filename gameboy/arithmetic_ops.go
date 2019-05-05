package gameboy

// makeADD creates an instruction that adds the value of reg, an 8-bit
// register, into register A.
func makeADD(reg register8) instruction {
	return func(state *State) instruction {
		result, carry, halfCarry := add(state.regA.get(), reg.get(), 0)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// addFromMemHL adds the value stored in the memory address specified by HL
// into register A.
func addFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Load from memory and do operation
		memVal := state.mmu.at(state.regHL.get())

		result, carry, halfCarry := add(state.regA.get(), memVal, 0)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// makeADDToHL creates an instruction that adds the value of the given 16-bit
// register into register HL.
func makeADDToHL(reg register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction and set the L register

		regLowerByte, _ := split16(reg.get())
		hlLowerByte, _ := split16(state.regHL.get())

		lVal, carry, _ := add(regLowerByte, hlLowerByte, 0)
		state.regL.set(lVal)

		return func(state *State) instruction {
			// M-Cycle 1: Set the H register

			_, regUpperByte := split16(reg.get())
			_, hlUpperByte := split16(state.regHL.get())

			hVal, carry, halfCarry := add(regUpperByte, hlUpperByte, carry)

			state.setHalfCarryFlag(halfCarry == 1)
			state.setCarryFlag(carry == 1)
			state.setSubtractFlag(false)

			state.regH.set(hVal)

			return nil
		}
	}
}

// add8BitImm loads an 8-bit immediate value and adds it to register A, storing
// the results in register A.
func add8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation

		imm := state.incrementPC()

		result, carry, halfCarry := add(state.regA.get(), imm, 0)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// addToSP loads an immediate signed 8-bit value and adds it to the stack
// pointer register.
func addToSP(state *State) instruction {
	// TODO(velovix): Implement this in an m-cycle accurate way
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value

		immUnsigned := state.incrementPC()
		imm := int8(immUnsigned)

		return func(state *State) instruction {
			// M-Cycle 2: Write to lower SP byte

			return func(state *State) instruction {
				// M-Cycle 3: Write to upper SP byte

				// This instruction's behavior for the carry and half carry flags is
				// very weird.  When checking for a carry and half carry, the immediate
				// value is treated as _unsigned_ for some reason and only the lowest 8
				// bits of the stack pointer are considered.
				lowerSP, _ := split16(state.regSP.get())

				state.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
				state.setCarryFlag(isCarry(lowerSP, immUnsigned))
				state.setZeroFlag(false)
				state.setSubtractFlag(false)

				state.regSP.set(uint16(int(state.regSP.get()) + int(imm)))

				return nil
			}
		}
	}
}

// makeADC creates an instruction that adds the value of the given register and
// the carry bit to register A, storing the results in register A.
//
// regA = regA + reg + carry bit
func makeADC(reg register8) instruction {
	return func(state *State) instruction {
		var carry uint8
		if state.getCarryFlag() {
			carry = 1
		}

		result, carry, halfCarry := add(state.regA.get(), reg.get(), carry)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// adcFromMemHL adds the value in memory at the address specified by register
// HL to register A, then adds the carry bit. Results are stored in register A.
//
// regA = regA + mem[regHL] + carry bit
func adcFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation
		memVal := state.mmu.at(state.regHL.get())

		var carry uint8
		if state.getCarryFlag() {
			carry = 1
		}

		result, carry, halfCarry := add(state.regA.get(), memVal, carry)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// adc8BitImm loads and 8-bit immediate value and adds it and the carry
// register to register A, storing the result in register A.
//
// regA = regA + imm + carry bit
func adc8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation
		imm := state.incrementPC()

		var carry uint8
		if state.getCarryFlag() {
			carry = 1
		}

		result, carry, halfCarry := add(state.regA.get(), imm, carry)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setCarryFlag(carry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)

		state.regA.set(result)

		return nil
	}
}

// makeSUB creates an instruction that subtracts the value of reg, an 8-bit
// register, from register A.
func makeSUB(reg register8) instruction {
	return func(state *State) instruction {
		result, borrow, halfBorrow := subtract(state.regA.get(), reg.get(), 0)

		state.setHalfCarryFlag(halfBorrow == 1)
		state.setCarryFlag(borrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(result)

		return nil
	}
}

// subFromMemHL subtracts the value in memory at the address specified by HL
// from register A.
func subFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation

		memVal := state.mmu.at(state.regHL.get())

		result, borrow, halfBorrow := subtract(state.regA.get(), memVal, 0)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(result)

		return nil
	}
}

// sub8BitImm loads an 8-bit immediate value and subtracts it from register A,
// storing the result in register A.
func sub8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation

		imm := state.incrementPC()

		result, borrow, halfBorrow := subtract(state.regA.get(), imm, 0)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(result)

		return nil
	}
}

// makeSBC creates an instruction that subtracts the value of the given
// register and the carry bit from register A, storing the results in register
// A.
//
// regA = regA - reg - carry bit
func makeSBC(reg register8) instruction {
	return func(state *State) instruction {
		var borrow uint8
		if state.getCarryFlag() {
			borrow = 1
		}

		result, borrow, halfBorrow := subtract(state.regA.get(), reg.get(), borrow)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(result)

		return nil
	}
}

// sbcFromMemHL subtracts the value in memory at the address specified by
// register HL to register A, then subtracts the carry bit. Results are stored
// in register A.
//
// regA = regA - mem[regHL] - carry bit
func sbcFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation

		memVal := state.mmu.at(state.regHL.get())

		var carry uint8
		if state.getCarryFlag() {
			carry = 1
		}

		result, borrow, halfBorrow := subtract(state.regA.get(), memVal, carry)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(state.regA.get() - memVal - carry)

		return nil
	}
}

// sbc8BitImm loads and 8-bit immediate value and subtracts it and the carry
// register from register A, storing the result in register A.
//
// regA = regA - imm - carry bit
func sbc8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Get immediate value and do operation

		imm := state.incrementPC()

		var carry uint8
		if state.getCarryFlag() {
			carry = 1
		}

		result, borrow, halfBorrow := subtract(state.regA.get(), imm, carry)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		state.regA.set(result)

		return nil
	}
}

// makeAND creates an instruction that performs a bitwise & on the given
// register and register A, storing the result in register A.
func makeAND(reg register8) instruction {
	return func(state *State) instruction {
		result := state.regA.get() & reg.get()

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// andFromMemHL performs a bitwise & on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func andFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation

		memVal := state.mmu.at(state.regHL.get())

		result := state.regA.get() & memVal

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// and8BitImm performs a bitwise & on register A and an immediate value,
// storing the result in register A.
func and8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Get immediate value and do operation

		imm := state.incrementPC()

		result := state.regA.get() & imm

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// makeOR creates an instruction that performs a bitwise | on the given
// register and register A, storing the result in register A.
func makeOR(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction and do operation

		result := state.regA.get() | reg.get()

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// orFromMemHL performs a bitwise | on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func orFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation

		memVal := state.mmu.at(state.regHL.get())

		result := state.regA.get() | memVal

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// or8BitImm performs a bitwise | on register A and an immediate value,
// storing the result in register A.
func or8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Load immediate value and do operation

		imm := state.incrementPC()

		result := state.regA.get() | imm

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// makeXOR creates an instruction that performs a bitwise ^ on register A and
// the given register, storing the result in register A.
func makeXOR(reg register8) instruction {
	return func(state *State) instruction {
		result := state.regA.get() ^ reg.get()

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// xorFromMemHL performs a bitwise ^ on the value in memory specified by
// register HL and register A, storing the result in register A.
func xorFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation

		memVal := state.mmu.at(state.regHL.get())

		result := state.regA.get() ^ memVal

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// xor8BitImm performs a bitwise ^ on register A and an immediate value,
// storing the result in register A.
func xor8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation

		imm := state.incrementPC()

		result := state.regA.get() ^ imm

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		state.regA.set(result)

		return nil
	}
}

// makeINC8Bit creates an instruction that increments the given 8-bit register
// by 1.
func makeINC8Bit(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Do operation

		result, _, halfCarry := add(reg.get(), 1, 0)

		state.setZeroFlag(result == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(halfCarry == 1)

		reg.set(result)

		return nil
	}
}

// makeINC16Bit creates an instruction that increments the given 16-bit
// register by 1.
func makeINC16Bit(reg register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Write to the lower register

		lowerResult, carry, _ := add(uint8(reg.get()), 1, 0)
		reg.setLower(lowerResult)

		return func(state *State) instruction {
			// M-Cycle 1: Write to the upper register

			_, upperVal := split16(reg.get())
			upperResult, _, _ := add(upperVal, 0, carry)
			reg.setUpper(upperResult)

			return nil
		}
	}
}

// incMemHL increments the value in memory at the address specified by register
// HL.
func incMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory

		addr := state.regHL.get()
		memVal := state.mmu.at(addr)

		return func(state *State) instruction {
			// M-Cycle 2: Do operation and write to memory

			result, _, halfCarry := add(memVal, 1, 0)

			state.setZeroFlag(result == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(halfCarry == 1)

			state.mmu.set(addr, result)

			return nil
		}
	}
}

// makeDEC8Bit creates an instruction that decrements the given 8-bit register
// by 1.
func makeDEC8Bit(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Do operation

		result, _, halfCarry := subtract(reg.get(), 1, 0)

		state.setHalfCarryFlag(halfCarry == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		reg.set(result)

		return nil
	}
}

// makeDEC16Bit creates an instruction that decrements the given 16-bit
// register by 1.
func makeDEC16Bit(reg register16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Write to lower byte

		lowerResult, borrow, _ := subtract(uint8(reg.get()), 1, 0)
		reg.setLower(lowerResult)

		return func(state *State) instruction {
			// M-Cycle 1: Write to upper byte

			_, upperVal := split16(reg.get())
			upperResult, _, _ := subtract(upperVal, 0, borrow)
			reg.setUpper(upperResult)

			return nil
		}
	}
}

// decMemHL decrements the value in memory at the address specified by register
// HL.
func decMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory

		addr := state.regHL.get()
		memVal := state.mmu.at(addr)

		return func(state *State) instruction {
			// M-Cycle 2: Do operation and write to memory

			result, _, halfBorrow := subtract(memVal, 1, 0)

			state.setZeroFlag(result == 0)
			state.setSubtractFlag(true)
			state.setHalfCarryFlag(halfBorrow == 1)

			state.mmu.set(addr, result)

			return nil
		}
	}
}

// makeCP creates an instruction that compares the value in register A with the
// value of the given register and sets flags accordingly. The semantics are
// the same as the SUB operator, but the result value is not saved.
func makeCP(reg register8) instruction {
	return func(state *State) instruction {
		result, borrow, halfBorrow := subtract(state.regA.get(), reg.get(), 0)

		state.setCarryFlag(borrow == 1)
		state.setHalfCarryFlag(halfBorrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		return nil
	}
}

// cpFromMemHL compares the value in register A with the value in memory at the
// address specified by the HL register and sets flags accordingly. The
// semantics are the same as the SUB operator, but the result value is not
// saved.
func cpFromMemHL(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read from memory and do operation
		memVal := state.mmu.at(state.regHL.get())

		result, borrow, halfBorrow := subtract(state.regA.get(), memVal, 0)

		state.setHalfCarryFlag(halfBorrow == 1)
		state.setCarryFlag(borrow == 1)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		return nil
	}
}

// cp8BitImm compares register A with an immediate value and sets flags
// accordingly. The semantics are the same as the SUB operator, but the result
// value is not saved.
func cp8BitImm(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read immediate value and do operation

		imm := state.incrementPC()

		result, borrow, halfBorrow := subtract(state.regA.get(), imm, 0)

		state.setHalfCarryFlag(halfBorrow == 0)
		state.setCarryFlag(borrow == 0)
		state.setZeroFlag(result == 0)
		state.setSubtractFlag(true)

		return nil
	}
}

// daa "corrects" the result of a previous add or subtract operation by
// reformatting it to fit the BCD format. It's a weird instruction.
//
// BCD stands for "binary coded decimal", and it's a special way to format
// numbers in binary that is based on decimal number formatting. Each decimal
// digit is separately represented by 4 bits.
//
// For example, the number 26 is represented as 0b11010 in normal binary, but
// in BCD it is represented as 0b00100110. If we split it up into 4-bit nibbles,
// we can see that the first nibble is 2 and the second nibble is 6.
//
// BCD:     0010  0110
// Decimal:    2     6
//
// Now, imagine that you have two BCD-encoded numbers and you add them together
// with an add instruction. The CPU doesn't know that these numbers are BCD and
// so it will add them together like they are normal binary numbers. The result
// will, as a result, no longer be in BCD format and will instead be some
// nonsense value.
//
// That's where this instruction comes in. This instruction makes the
// necessary corrections to the result of the last operation to make it once
// again BCD encoded. If you're doing math with BCD numbers, this instruction
// would be called after every add or subtract instruction. But why on earth
// would you?
func daa(state *State) instruction {
	aVal := state.regA.get()

	var correction uint8

	// Check if there was a half borrow or if the least significant nibble
	// (which represents the first decimal digit) is overfilled. If it is, a
	// correction needs to be performed.
	if state.getHalfCarryFlag() || (!state.getSubtractFlag() && (aVal&0xF) > 9) {
		correction = 0x06
	}

	// Check if there was a full borrow or if the most significant nibble
	// (which represents the second decimal digit) is overfilled (I think?). If
	// it is, a correction needs to be performed.
	if state.getCarryFlag() || (!state.getSubtractFlag() && aVal > 0x99) {
		correction |= 0x60
		state.setCarryFlag(true)
	} else {
		state.setCarryFlag(false)
	}

	// The direction of the correction depends on what the last operation was
	if state.getSubtractFlag() {
		aVal = state.regA.set(aVal - correction)
	} else {
		aVal = state.regA.set(aVal + correction)
	}

	state.setZeroFlag(aVal == 0)
	state.setHalfCarryFlag(false)

	return nil
}
