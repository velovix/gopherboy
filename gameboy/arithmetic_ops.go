package gameboy

// makeADD creates an instruction that adds the value of reg, an 8-bit
// register, into register A.
func makeADD(reg register8) instruction {
	return func(state *State) int {
		state.setHalfCarryFlag(isHalfCarry(state.regA.get(), reg.get()))
		state.setCarryFlag(isCarry(state.regA.get(), reg.get()))

		state.regA.set(state.regA.get() + reg.get())

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(false)

		return 4
	}
}

// addFromMemHL adds the value stored in the memory address specified by HL
// into register A.
func addFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())

	state.setHalfCarryFlag(isHalfCarry(state.regA.get(), memVal))
	state.setCarryFlag(isCarry(state.regA.get(), memVal))

	state.regA.set(state.regA.get() + memVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)

	return 8
}

// makeADDToHL creates an instruction that adds the value of the given 16-bit
// register into register HL.
func makeADDToHL(reg register16) instruction {
	return func(state *State) int {
		state.setHalfCarryFlag(isHalfCarry16(state.regHL.get(), reg.get()))
		state.setCarryFlag(isCarry16(state.regHL.get(), reg.get()))
		state.setSubtractFlag(false)

		state.regHL.set(state.regHL.get() + reg.get())

		return 8
	}
}

// add8BitImm loads an 8-bit immediate value and adds it to register A, storing
// the results in register A.
func add8BitImm(state *State) int {
	imm := state.incrementPC()

	state.setHalfCarryFlag(isHalfCarry(state.regA.get(), imm))
	state.setCarryFlag(isCarry(state.regA.get(), imm))

	state.regA.set(imm + state.regA.get())

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)

	return 8
}

// addToSP loads an immediate signed 8-bit value and adds it to the stack
// pointer register.
func addToSP(state *State) int {
	immUnsigned := state.incrementPC()
	imm := int8(immUnsigned)

	// This instruction's behavior for the carry and half carry flags is very
	// weird.
	//
	// When checking for a carry and half carry, the immediate value is treated
	// as _unsigned_ for some reason and only the lowest 8 bits of the stack
	// pointer are considered.
	lowerSP, _ := split16(state.regSP.get())
	state.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
	state.setCarryFlag(isCarry(lowerSP, immUnsigned))

	state.regSP.set(uint16(int(state.regSP.get()) + int(imm)))

	state.setZeroFlag(false)
	state.setSubtractFlag(false)

	return 16
}

// makeADC creates an instruction that adds the value of the given register and
// the carry bit to register A, storing the results in register A.
//
// regA = regA + reg + carry bit
func makeADC(reg register8) instruction {
	return func(state *State) int {
		carryVal := uint8(0)

		if state.getCarryFlag() {
			carryVal = 1
		}

		state.setHalfCarryFlag(isHalfCarry(state.regA.get(), reg.get(), carryVal))
		state.setCarryFlag(isCarry(state.regA.get(), reg.get(), carryVal))

		state.regA.set(state.regA.get() + reg.get() + carryVal)

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(false)

		return 4
	}
}

// adcFromMemHL adds the value in memory at the address specified by register
// HL to register A, then adds the carry bit. Results are stored in register A.
//
// regA = regA + mem[regHL] + carry bit
func adcFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setHalfCarryFlag(isHalfCarry(state.regA.get(), memVal, carryVal))
	state.setCarryFlag(isCarry(state.regA.get(), memVal, carryVal))

	state.regA.set(state.regA.get() + memVal + carryVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)

	return 8
}

// adc8BitImm loads and 8-bit immediate value and adds it and the carry
// register to register A, storing the result in register A.
//
// regA = regA + imm + carry bit
func adc8BitImm(state *State) int {
	imm := state.incrementPC()
	var carry uint8

	if state.getCarryFlag() {
		carry = 1
	}

	state.setHalfCarryFlag(isHalfCarry(state.regA.get(), imm, carry))
	state.setCarryFlag(isCarry(state.regA.get(), imm, carry))

	state.regA.set(state.regA.get() + imm + carry)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)

	return 8
}

// makeSUB creates an instruction that subtracts the value of reg, an 8-bit
// register, from register A.
func makeSUB(reg register8) instruction {
	return func(state *State) int {
		// A carry occurs if the value we're subtracting is greater than register
		// A, meaning that the register A value rolled over
		state.setCarryFlag(reg.get() > state.regA.get())
		state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), reg.get()))

		state.regA.set(state.regA.get() - reg.get())

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(true)

		return 4
	}
}

// subFromMemHL subtracts the value in memory at the address specified by HL
// from register A.
func subFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(memVal > state.regA.get())
	state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), memVal))

	state.regA.set(state.regA.get() - memVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(true)

	return 8
}

// sub8BitImm loads an 8-bit immediate value and subtracts it from register A,
// storing the result in register A.
func sub8BitImm(state *State) int {
	imm := state.incrementPC()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(imm > state.regA.get())
	state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), imm))

	state.regA.set(state.regA.get() - imm)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(true)

	return 8
}

// makeSBC creates an instruction that subtracts the value of the given
// register and the carry bit from register A, storing the results in register
// A.
//
// regA = regA - reg - carry bit
func makeSBC(reg register8) instruction {
	return func(state *State) int {
		carryVal := uint8(0)

		if state.getCarryFlag() {
			carryVal = 1
		}

		state.setCarryFlag(isBorrow(state.regA.get(), reg.get(), carryVal))
		state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), reg.get(), carryVal))

		state.regA.set(state.regA.get() - reg.get() - carryVal)

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(true)

		return 4
	}
}

// sbcFromMemHL subtracts the value in memory at the address specified by
// register HL to register A, then subtracts the carry bit. Results are stored
// in register A.
//
// regA = regA - mem[regHL] - carry bit
func sbcFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setCarryFlag(isBorrow(state.regA.get(), memVal, carryVal))
	state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), memVal, carryVal))

	state.regA.set(state.regA.get() - memVal - carryVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(true)

	return 8
}

// sbc8BitImm loads and 8-bit immediate value and subtracts it and the carry
// register from register A, storing the result in register A.
//
// regA = regA - imm - carry bit
func sbc8BitImm(state *State) int {
	imm := state.incrementPC()
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setCarryFlag(isBorrow(state.regA.get(), imm, carryVal))
	state.setHalfCarryFlag(isHalfBorrow(state.regA.get(), imm, carryVal))

	state.regA.set(state.regA.get() - imm - carryVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(true)

	return 8
}

// makeAND creates an instruction that performs a bitwise & on the given
// register and register A, storing the result in register A.
func makeAND(reg register8) instruction {
	return func(state *State) int {
		state.regA.set(state.regA.get() & reg.get())

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)
		state.setCarryFlag(false)

		return 4
	}
}

// andFromMemHL performs a bitwise & on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func andFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())

	state.regA.set(state.regA.get() & memVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(true)
	state.setCarryFlag(false)

	return 8
}

// and8BitImm performs a bitwise & on register A and an immediate value,
// storing the result in register A.
func and8BitImm(state *State) int {
	imm := state.incrementPC()

	state.regA.set(state.regA.get() & imm)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(true)
	state.setCarryFlag(false)

	return 8
}

// makeOR creates an instruction that performs a bitwise | on the given
// register and register A, storing the result in register A.
func makeOR(reg register8) instruction {
	return func(state *State) int {
		state.regA.set(state.regA.get() | reg.get())

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		return 4
	}
}

// orFromMemHL performs a bitwise | on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func orFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())

	state.regA.set(state.regA.get() | memVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	return 8
}

// or8BitImm performs a bitwise | on register A and an immediate value,
// storing the result in register A.
func or8BitImm(state *State) int {
	imm := state.incrementPC()

	state.regA.set(state.regA.get() | imm)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	return 8
}

// makeXOR creates an instruction that performs a bitwise ^ on register A and
// the given register, storing the result in register A.
func makeXOR(reg register8) instruction {
	return func(state *State) int {
		state.regA.set(state.regA.get() ^ reg.get())

		state.setZeroFlag(state.regA.get() == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		return 4
	}
}

// xorFromMemHL performs a bitwise ^ on the value in memory specified by
// register HL and register A, storing the result in register A.
func xorFromMemHL(state *State) int {
	memVal := state.mmu.at(state.regHL.get())

	state.regA.set(state.regA.get() ^ memVal)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	return 8
}

// xor8BitImm performs a bitwise ^ on register A and an immediate value,
// storing the result in register A.
func xor8BitImm(state *State) int {
	imm := state.incrementPC()

	state.regA.set(state.regA.get() ^ imm)

	state.setZeroFlag(state.regA.get() == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	return 8
}

// makeINC8Bit creates an instruction that increments the given 8-bit register
// by 1.
func makeINC8Bit(reg register8) instruction {
	return func(state *State) int {
		oldVal := reg.get()
		newVal := reg.set(oldVal + 1)

		state.setZeroFlag(newVal == 0)
		state.setSubtractFlag(false)
		// A half carry occurs only if the bottom 4 bits of the number are 1,
		// meaning all those "slots" are "filled"
		state.setHalfCarryFlag(oldVal&0x0F == 0x0F)

		return 4
	}
}

// makeINC16Bit creates an instruction that increments the given 16-bit
// register by 1.
func makeINC16Bit(reg register16) instruction {
	return func(state *State) int {
		reg.set(reg.get() + 1)

		return 8
	}
}

// incMemHL increments the value in memory at the address specified by register
// HL.
func incMemHL(state *State) int {
	addr := state.regHL.get()

	oldVal := state.mmu.at(addr)
	state.mmu.set(addr, oldVal+1)
	newVal := state.mmu.at(addr)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	state.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	return 12
}

// makeDEC8Bit creates an instruction that decrements the given 8-bit register
// by 1.
func makeDEC8Bit(reg register8) instruction {
	return func(state *State) int {
		oldVal := reg.get()

		newVal := reg.set(oldVal - 1)

		state.setHalfCarryFlag(isHalfBorrow(oldVal, 1))
		state.setZeroFlag(newVal == 0)
		state.setSubtractFlag(true)

		return 4
	}
}

// makeDEC16Bit creates an instruction that decrements the given 16-bit
// register by 1.
func makeDEC16Bit(reg register16) instruction {
	return func(state *State) int {
		reg.set(reg.get() - 1)

		return 8
	}
}

// decMemHL decrements the value in memory at the address specified by register
// HL.
func decMemHL(state *State) int {
	addr := state.regHL.get()

	oldVal := state.mmu.at(addr)

	state.mmu.set(addr, oldVal-1)
	newVal := state.mmu.at(addr)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(true)
	state.setHalfCarryFlag(isHalfBorrow(oldVal, 1))

	return 12
}

// makeCP creates an instruction that compares the value in register A with the
// value of the given register and sets flags accordingly. The semantics are
// the same as the SUB operator, but the result value is not saved.
func makeCP(reg register8) instruction {
	return func(state *State) int {
		aVal := state.regA.get()
		regVal := reg.get()

		// A carry occurs if the value we're subtracting is greater than register
		// A, meaning that the register A value rolled over
		state.setCarryFlag(regVal > aVal)
		state.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

		subVal := aVal - regVal

		state.setZeroFlag(subVal == 0)
		state.setSubtractFlag(true)

		return 4
	}
}

// cpFromMemHL compares the value in register A with the value in memory at the
// address specified by the HL register and sets flags accordingly. The
// semantics are the same as the SUB operator, but the result value is not
// saved.
func cpFromMemHL(state *State) int {
	aVal := state.regA.get()
	memVal := state.mmu.at(state.regHL.get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(memVal > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	subVal := aVal - memVal

	state.setZeroFlag(subVal == 0)
	state.setSubtractFlag(true)

	return 8
}

// cp8BitImm compares register A with an immediate value and sets flags
// accordingly. The semantics are the same as the SUB operator, but the result
// value is not saved.
func cp8BitImm(state *State) int {
	aVal := state.regA.get()
	imm := state.incrementPC()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(aVal < imm)
	state.setHalfCarryFlag(isHalfBorrow(aVal, imm))

	subVal := aVal - imm

	state.setZeroFlag(subVal == 0)
	state.setSubtractFlag(true)

	return 8
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
// will, as a result, be incorrect from a BCD perspective.
//
// That's where this instructions comes in. This instruction makes the
// necessary corrections to the result of the last operation to make it once
// again BCD encoded. If you're doing math with BCD numbers, this instruction
// would be called after every add or subtract instruction.
func daa(state *State) int {
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

	return 4
}
