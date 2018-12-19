package main

import "fmt"

// add adds the value of reg, an 8-bit register, into register A.
func add(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	state.setHalfCarryFlag(isHalfCarry(aVal, regVal))
	state.setCarryFlag(isCarry(aVal, regVal))

	aVal = state.regs8[regA].set(aVal + regVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD A,%v\n", reg)
	}
	return 4
}

// addFromMemHL adds the value stored in the memory address specified by HL
// into register A.
func addFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	hlVal := state.regs16[regHL].get()
	memVal := state.mmu.at(hlVal)

	state.setHalfCarryFlag(isHalfCarry(aVal, memVal))
	state.setCarryFlag(isCarry(aVal, memVal))

	aVal = state.regs8[regA].set(aVal + memVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD A,(HL)\n")
	}
	return 8
}

// addToHL adds the value of the given 16-bit register into register HL.
func addToHL(state *State, reg registerType) int {
	hlVal := state.regs16[regHL].get()
	regVal := state.regs16[reg].get()

	state.setHalfCarryFlag(isHalfCarry16(hlVal, regVal))
	state.setCarryFlag(isCarry16(hlVal, regVal))
	state.setSubtractFlag(false)

	hlVal = state.regs16[regHL].set(hlVal + regVal)

	if printInstructions {
		fmt.Printf("ADD HL,%v\n", reg)
	}
	return 8
}

// add8BitImm loads an 8-bit immediate value and adds it to register A, storing
// the results in register A.
func add8BitImm(state *State) int {
	imm := state.incrementPC()
	aVal := state.regs8[regA].get()

	state.setHalfCarryFlag(isHalfCarry(aVal, imm))
	state.setCarryFlag(isCarry(aVal, imm))

	aVal = state.regs8[regA].set(imm + aVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD %v,%#x\n", regA, imm)
	}
	return 8
}

// addToSP loads an immediate signed 8-bit value and adds it to the stack
// pointer register.
func addToSP(state *State) int {
	immUnsigned := state.incrementPC()
	imm := int8(immUnsigned)
	spVal := state.regs16[regSP].get()

	// This instruction's behavior for the carry and half carry flags is very
	// weird.
	//
	// When checking for a carry and half carry, the immediate value is treated
	// as _unsigned_ for some reason and only the lowest 8 bits of the stack
	// pointer are considered.
	lowerSP, _ := split16(spVal)
	state.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
	state.setCarryFlag(isCarry(lowerSP, immUnsigned))

	state.regs16[regSP].set(uint16(int(spVal) + int(imm)))

	state.setZeroFlag(false)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD SP,%#x\n", imm)
	}
	return 16
}

// adc adds the value of the given register and the carry bit to register A,
// storing the results in register A.
//
// regA = regA + reg + carry bit
func adc(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setHalfCarryFlag(isHalfCarry(aVal, regVal, carryVal))
	state.setCarryFlag(isCarry(aVal, regVal, carryVal))

	aVal = state.regs8[regA].set(aVal + regVal + carryVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC A,%v\n", reg)
	}
	return 4
}

// adcFromMemHL adds the value in memory at the address specified by register
// HL to register A, then adds the carry bit. Results are stored in register A.
//
// regA = regA + mem[regHL] + carry bit
func adcFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setHalfCarryFlag(isHalfCarry(aVal, memVal, carryVal))
	state.setCarryFlag(isCarry(aVal, memVal, carryVal))

	aVal = state.regs8[regA].set(aVal + memVal + carryVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC %v,(%v)\n", regA, regHL)
	}
	return 4
}

// adc8BitImm loads and 8-bit immediate value and adds it and the carry
// register to register A, storing the result in register A.
//
// regA = regA + imm + carry bit
func adc8BitImm(state *State) int {
	aVal := state.regs8[regA].get()
	imm := state.incrementPC()
	var carry uint8

	if state.getCarryFlag() {
		carry = 1
	}

	state.setHalfCarryFlag(isHalfCarry(aVal, imm, carry))
	state.setCarryFlag(isCarry(aVal, imm, carry))

	aVal = state.regs8[regA].set(aVal + imm + carry)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC A,%#x\n", imm)
	}
	return 4
}

// sub subtracts the value of reg, an 8-bit register, from register A.
func sub(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(regVal > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

	aVal = state.regs8[regA].set(aVal - regVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB %v\n", reg)
	}
	return 4
}

// subFromMemHL subtracts the value in memory at the address specified by HL
// from register A.
func subFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(memVal > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	aVal = state.regs8[regA].set(aVal - memVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB (%v)\n", regHL)
	}
	return 8
}

// sub8BitImm loads an 8-bit immediate value and subtracts it from register A,
// storing the result in register A.
func sub8BitImm(state *State) int {
	imm := state.incrementPC()
	aVal := state.regs8[regA].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(imm > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, imm))

	aVal = state.regs8[regA].set(aVal - imm)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB %#x\n", imm)
	}
	return 8
}

// sbc subtracts the value of the given register and the carry bit from
// register A, storing the results in register A.
//
// regA = regA - reg - carry bit
func sbc(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setCarryFlag(isBorrow(aVal, regVal, carryVal))
	state.setHalfCarryFlag(isHalfBorrow(aVal, regVal, carryVal))

	aVal = state.regs8[regA].set(aVal - regVal - carryVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SBC %v\n", reg)
	}
	return 4
}

// sbcFromMemHL subtracts the value in memory at the address specified by
// register HL to register A, then subtracts the carry bit. Results are stored
// in register A.
//
// regA = regA - mem[regHL] - carry bit
func sbcFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setCarryFlag(isBorrow(aVal, memVal, carryVal))
	state.setHalfCarryFlag(isHalfBorrow(aVal, memVal, carryVal))

	aVal = state.regs8[regA].set(aVal - memVal - carryVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SBC (%v)\n", regHL)
	}
	return 8
}

// sbc8BitImm loads and 8-bit immediate value and subtracts it and the carry
// register from register A, storing the result in register A.
//
// regA = regA - imm - carry bit
func sbc8BitImm(state *State) int {
	aVal := state.regs8[regA].get()
	imm := state.incrementPC()
	carryVal := uint8(0)

	if state.getCarryFlag() {
		carryVal = 1
	}

	state.setCarryFlag(isBorrow(aVal, imm, carryVal))
	state.setHalfCarryFlag(isHalfBorrow(aVal, imm, carryVal))

	aVal = state.regs8[regA].set(aVal - imm - carryVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SBC %v,%#x\n", regA, imm)
	}
	return 8
}

// and performs a bitwise & on the given register and register A, storing the
// result in register A.
func and(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	aVal = state.regs8[regA].set(aVal & regVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(true)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND %v\n", reg)
	}
	return 4
}

// andFromMemHL performs a bitwise & on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func andFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())

	aVal = state.regs8[regA].set(aVal & memVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(true)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND (%v)\n", regHL)
	}
	return 8
}

// and8BitImm performs a bitwise & on register A and an immediate value,
// storing the result in register A.
func and8BitImm(state *State) int {
	imm := state.incrementPC()
	aVal := state.regs8[regA].get()

	aVal = state.regs8[regA].set(aVal & imm)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(true)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND %#x\n", imm)
	}
	return 8
}

// or performs a bitwise | on the given register and register A, storing the
// result in register A.
func or(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	aVal = state.regs8[regA].set(aVal | regVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR %v\n", reg)
	}
	return 4
}

// orFromMemHL performs a bitwise | on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func orFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())

	aVal = state.regs8[regA].set(aVal | memVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR (%v)\n", regHL)
	}
	return 8
}

// or8BitImm performs a bitwise | on register A and an immediate value,
// storing the result in register A.
func or8BitImm(state *State) int {
	aVal := state.regs8[regA].get()
	imm := state.incrementPC()

	aVal = state.regs8[regA].set(aVal | imm)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR %#x\n", imm)
	}
	return 8
}

// xor performs a bitwise ^ on register A and the given register, storing the
// result in register A.
func xor(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	aVal = state.regs8[regA].set(aVal ^ regVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR %v\n", reg)
	}
	return 4
}

// xorFromMemHL performs a bitwise ^ on the value in memory specified by
// register HL and register A, storing the result in register A.
func xorFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())

	aVal = state.regs8[regA].set(aVal ^ memVal)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR (%v)\n", regHL)
	}
	return 8
}

// xor8BitImm performs a bitwise ^ on register A and an immediate value,
// storing the result in register A.
func xor8BitImm(state *State) int {
	imm := state.incrementPC()
	aVal := state.regs8[regA].get()

	aVal = state.regs8[regA].set(aVal ^ imm)

	state.setZeroFlag(aVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR %#x\n", imm)
	}
	return 8
}

// inc8Bit increments the given 8-bit register by 1.
func inc8Bit(state *State, reg registerType) int {
	oldVal := state.regs8[reg].get()
	newVal := state.regs8[reg].set(oldVal + 1)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	state.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	if printInstructions {
		fmt.Printf("INC %v\n", reg)
	}
	return 4
}

// inc16Bit increments the given 16-bit register by 1.
func inc16Bit(state *State, reg registerType) int {
	oldVal := state.regs16[reg].get()

	state.regs16[reg].set(oldVal + 1)

	if printInstructions {
		fmt.Printf("INC %v\n", reg)
	}
	return 8
}

// incMemHL increments the value in memory at the address specified by register
// HL.
func incMemHL(state *State) int {
	addr := state.regs16[regHL].get()

	oldVal := state.mmu.at(addr)
	state.mmu.set(addr, oldVal+1)
	newVal := state.mmu.at(addr)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	state.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	if printInstructions {
		fmt.Printf("INC (HL)\n")
	}

	return 12
}

// dec8Bit decrements the given 8-bit register by 1.
func dec8Bit(state *State, reg registerType) int {
	oldVal := state.regs8[reg].get()

	newVal := state.regs8[reg].set(oldVal - 1)

	state.setHalfCarryFlag(isHalfBorrow(oldVal, 1))
	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("DEC %v\n", reg)
	}
	return 4
}

// dec16Bit decrements the given 16-bit register by 1.
func dec16Bit(state *State, reg registerType) int {
	state.regs16[reg].set(state.regs16[reg].get() - 1)

	if printInstructions {
		fmt.Printf("DEC %v\n", reg)
	}
	return 8
}

// decMemHL decrements the value in memory at the address specified by register
// HL.
func decMemHL(state *State) int {
	addr := state.regs16[regHL].get()

	oldVal := state.mmu.at(addr)

	state.mmu.set(addr, oldVal-1)
	newVal := state.mmu.at(addr)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(true)
	state.setHalfCarryFlag(isHalfBorrow(oldVal, 1))

	if printInstructions {
		fmt.Printf("DEC (HL)\n")
	}

	return 12
}

// cp compares the value in register A with the value of the given register and
// sets flags accordingly.  The semantics are the same as the SUB operator, but
// the result value is not saved.
func cp(state *State, reg registerType) int {
	aVal := state.regs8[regA].get()
	regVal := state.regs8[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(regVal > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

	subVal := aVal - regVal

	state.setZeroFlag(subVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CP %v (%#x,%#x)\n",
			reg,
			state.regs8[regA].get(),
			state.regs8[reg].get())
	}
	return 4
}

// cpFromMemHL compares the value in register A with the value in memory at the
// address specified by the HL register and sets flags accordingly. The
// semantics are the same as the SUB operator, but the result value is not
// saved.
func cpFromMemHL(state *State) int {
	aVal := state.regs8[regA].get()
	memVal := state.mmu.at(state.regs16[regHL].get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(memVal > aVal)
	state.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	subVal := aVal - memVal

	state.setZeroFlag(subVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CP (%v)\n", regHL)
	}
	return 8
}

// cp8BitImm compares register A with an immediate value and sets flags
// accordingly. The semantics are the same as the SUB operator, but the result
// value is not saved.
func cp8BitImm(state *State) int {
	aVal := state.regs8[regA].get()
	imm := state.incrementPC()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	state.setCarryFlag(aVal < imm)
	state.setHalfCarryFlag(isHalfBorrow(aVal, imm))

	subVal := aVal - imm

	state.setZeroFlag(subVal == 0)
	state.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CP %#x\n", imm)
	}
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
	aVal := state.regs8[regA].get()

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
		aVal = state.regs8[regA].set(aVal - correction)
	} else {
		aVal = state.regs8[regA].set(aVal + correction)
	}

	state.setZeroFlag(aVal == 0)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("DAA\n")
	}

	return 4
}
