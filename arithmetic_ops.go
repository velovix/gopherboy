package main

import "fmt"

// add adds the value of reg, an 8-bit register, into register A.
func add(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	env.setHalfCarryFlag(isHalfCarry(aVal, regVal))
	env.setCarryFlag(isCarry(aVal, regVal))

	aVal = env.regs8[regA].set(aVal + regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD A,%v\n", reg)
	}
	return 4
}

// addFromMemHL adds the value stored in the memory address specified by HL
// into register A.
func addFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	hlVal := env.regs16[regHL].get()
	memVal := env.mmu.at(hlVal)

	env.setHalfCarryFlag(isHalfCarry(aVal, memVal))
	env.setCarryFlag(isCarry(aVal, memVal))

	aVal = env.regs8[regA].set(aVal + memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD A,(HL)\n")
	}
	return 8
}

// addToHL adds the value of the given 16-bit register into register HL.
func addToHL(env *environment, reg registerType) int {
	hlVal := env.regs16[regHL].get()
	regVal := env.regs16[reg].get()

	env.setHalfCarryFlag(isHalfCarry16(hlVal, regVal))
	env.setCarryFlag(isCarry16(hlVal, regVal))
	env.setSubtractFlag(false)

	hlVal = env.regs16[regHL].set(hlVal + regVal)

	if printInstructions {
		fmt.Printf("ADD HL,%v\n", reg)
	}
	return 8
}

// add8BitImm loads an 8-bit immediate value and adds it to register A, storing
// the results in register A.
func add8BitImm(env *environment) int {
	imm := env.incrementPC()
	aVal := env.regs8[regA].get()

	env.setHalfCarryFlag(isHalfCarry(aVal, imm))
	env.setCarryFlag(isCarry(aVal, imm))

	aVal = env.regs8[regA].set(imm + aVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD %v,%#x\n", regA, imm)
	}
	return 8
}

// addToSP loads an immediate signed 8-bit value and adds it to the stack
// pointer register.
func addToSP(env *environment) int {
	immUnsigned := env.incrementPC()
	imm := int8(immUnsigned)
	spVal := env.regs16[regSP].get()

	// This instruction's behavior for the carry and half carry flags is very
	// weird.
	//
	// When checking for a carry and half carry, the immediate value is treated
	// as _unsigned_ for some reason and only the lowest 8 bits of the stack
	// pointer are considered.
	lowerSP, _ := split16(spVal)
	env.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
	env.setCarryFlag(isCarry(lowerSP, immUnsigned))

	env.regs16[regSP].set(uint16(int(spVal) + int(imm)))

	env.setZeroFlag(false)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADD SP,%#x\n", imm)
	}
	return 16
}

// adc adds the value of the given register and the carry bit to register A,
// storing the results in register A.
//
// regA = regA + reg + carry bit
func adc(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	if env.getCarryFlag() {
		regVal++
	}

	env.setHalfCarryFlag(isHalfCarry(aVal, regVal))
	env.setCarryFlag(isCarry(aVal, regVal))

	aVal = env.regs8[regA].set(aVal + regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC A,%v\n", reg)
	}
	return 4
}

// adcFromMemHL adds the value in memory at the address specified by register
// HL to register A, then adds the carry bit. Results are stored in register A.
//
// regA = regA + mem[regHL] + carry bit
func adcFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	if env.getCarryFlag() {
		memVal++
	}

	env.setHalfCarryFlag(isHalfCarry(aVal, memVal))
	env.setCarryFlag(isCarry(aVal, memVal))

	aVal = env.regs8[regA].set(aVal + memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC %v,(%v)\n", regA, regHL)
	}
	return 4
}

// adc8BitImm loads and 8-bit immediate value and adds it and the carry
// register to register A, storing the result in register A.
//
// regA = regA + imm + carry bit
func adc8BitImm(env *environment) int {
	aVal := env.regs8[regA].get()
	imm := env.incrementPC()
	var carry uint8

	if env.getCarryFlag() {
		carry = 1
	}

	env.setHalfCarryFlag(isHalfCarry(aVal, imm+carry))
	env.setCarryFlag(isCarry(aVal, imm+carry))

	aVal = env.regs8[regA].set(aVal + imm + carry)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("ADC A,%#x\n", imm)
	}
	return 4
}

// sub subtracts the value of reg, an 8-bit register, from register A.
func sub(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(regVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

	aVal = env.regs8[regA].set(aVal - regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB %v\n", reg)
	}
	return 4
}

// subFromMemHL subtracts the value in memory at the address specified by HL
// from register A.
func subFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(memVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	aVal = env.regs8[regA].set(aVal - memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB (%v)\n", regHL)
	}
	return 8
}

// sub8BitImm loads an 8-bit immediate value and subtracts it from register A,
// storing the result in register A.
func sub8BitImm(env *environment) int {
	imm := env.incrementPC()
	aVal := env.regs8[regA].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(imm > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, imm))

	aVal = env.regs8[regA].set(aVal - imm)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SUB %#x\n", imm)
	}
	return 8
}

// sbc subtracts the value of the given register and the carry bit from
// register A, storing the results in register A.
//
// regA = regA - reg - carry bit
func sbc(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	if env.getCarryFlag() {
		regVal++
	}

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(regVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

	aVal = env.regs8[regA].set(aVal - regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

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
func sbcFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	if env.getCarryFlag() {
		memVal++
	}

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(memVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	aVal = env.regs8[regA].set(aVal - memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SBC (%v)\n", regHL)
	}
	return 8
}

// sbc8BitImm loads and 8-bit immediate value and subtracts it and the carry
// register from register A, storing the result in register A.
//
// regA = regA - imm - carry bit
func sbc8BitImm(env *environment) int {
	aVal := env.regs8[regA].get()
	imm := env.incrementPC()
	var carry uint8

	if env.getCarryFlag() {
		carry = 1
	}

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(imm+carry > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, imm+carry))

	aVal = env.regs8[regA].set(aVal - imm - carry)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("SBC %v,%#x\n", regA, imm)
	}
	return 8
}

// and performs a bitwise & on the given register and register A, storing the
// result in register A.
func and(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	aVal = env.regs8[regA].set(aVal & regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND %v\n", reg)
	}
	return 4
}

// andFromMemHL performs a bitwise & on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func andFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	aVal = env.regs8[regA].set(aVal & memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND (%v)\n", regHL)
	}
	return 8
}

// and8BitImm performs a bitwise & on register A and an immediate value,
// storing the result in register A.
func and8BitImm(env *environment) int {
	imm := env.incrementPC()
	aVal := env.regs8[regA].get()

	aVal = env.regs8[regA].set(aVal & imm)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("AND %#x\n", imm)
	}
	return 8
}

// or performs a bitwise | on the given register and register A, storing the
// result in register A.
func or(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	aVal = env.regs8[regA].set(aVal | regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR %v\n", reg)
	}
	return 4
}

// orFromMemHL performs a bitwise | on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func orFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	aVal = env.regs8[regA].set(aVal | memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR (%v)\n", regHL)
	}
	return 8
}

// or8BitImm performs a bitwise | on register A and an immediate value,
// storing the result in register A.
func or8BitImm(env *environment) int {
	aVal := env.regs8[regA].get()
	imm := env.incrementPC()

	aVal = env.regs8[regA].set(aVal | imm)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("OR %#x\n", imm)
	}
	return 8
}

// xor performs a bitwise ^ on register A and the given register, storing the
// result in register A.
func xor(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	aVal = env.regs8[regA].set(aVal ^ regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR %v\n", reg)
	}
	return 4
}

// xorFromMemHL performs a bitwise ^ on the value in memory specified by
// register HL and register A, storing the result in register A.
func xorFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	aVal = env.regs8[regA].set(aVal ^ memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR (%v)\n", regHL)
	}
	return 8
}

// xor8BitImm performs a bitwise ^ on register A and an immediate value,
// storing the result in register A.
func xor8BitImm(env *environment) int {
	imm := env.incrementPC()
	aVal := env.regs8[regA].get()

	aVal = env.regs8[regA].set(aVal ^ imm)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("XOR %#x\n", imm)
	}
	return 8
}

// inc8Bit increments the given 8-bit register by 1.
func inc8Bit(env *environment, reg registerType) int {
	oldVal := env.regs8[reg].get()
	newVal := env.regs8[reg].set(oldVal + 1)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	env.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	if printInstructions {
		fmt.Printf("INC %v\n", reg)
	}
	return 4
}

// inc16Bit increments the given 16-bit register by 1.
func inc16Bit(env *environment, reg registerType) int {
	oldVal := env.regs16[reg].get()

	env.regs16[reg].set(oldVal + 1)

	if printInstructions {
		fmt.Printf("INC %v\n", reg)
	}
	return 8
}

// incMemHL increments the value in memory at the address specified by register
// HL.
func incMemHL(env *environment) int {
	addr := env.regs16[regHL].get()

	oldVal := env.mmu.at(addr)
	env.mmu.set(addr, oldVal+1)
	newVal := env.mmu.at(addr)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	env.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	if printInstructions {
		fmt.Printf("INC (HL)\n")
	}

	return 12
}

// dec8Bit decrements the given 8-bit register by 1.
func dec8Bit(env *environment, reg registerType) int {
	oldVal := env.regs8[reg].get()

	newVal := env.regs8[reg].set(oldVal - 1)

	env.setHalfCarryFlag(isHalfBorrow(oldVal, 1))
	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("DEC %v\n", reg)
	}
	return 4
}

// dec16Bit decrements the given 16-bit register by 1.
func dec16Bit(env *environment, reg registerType) int {
	env.regs16[reg].set(env.regs16[reg].get() - 1)

	if printInstructions {
		fmt.Printf("DEC %v\n", reg)
	}
	return 8
}

// decMemHL decrements the value in memory at the address specified by register
// HL.
func decMemHL(env *environment) int {
	addr := env.regs16[regHL].get()

	oldVal := env.mmu.at(addr)

	env.mmu.set(addr, oldVal-1)
	newVal := env.mmu.at(addr)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(true)
	env.setHalfCarryFlag(isHalfBorrow(oldVal, 1))

	if printInstructions {
		fmt.Printf("DEC (HL)\n")
	}

	return 12
}

// cp compares the value in register A with the value of the given register and
// sets flags accordingly.  The semantics are the same as the SUB operator, but
// the result value is not saved.
func cp(env *environment, reg registerType) int {
	aVal := env.regs8[regA].get()
	regVal := env.regs8[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(regVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, regVal))

	subVal := aVal - regVal

	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CP %v (%#x,%#x)\n",
			reg,
			env.regs8[regA].get(),
			env.regs8[reg].get())
	}
	return 4
}

// cpFromMemHL compares the value in register A with the value in memory at the
// address specified by the HL register and sets flags accordingly. The
// semantics are the same as the SUB operator, but the result value is not
// saved.
func cpFromMemHL(env *environment) int {
	aVal := env.regs8[regA].get()
	memVal := env.mmu.at(env.regs16[regHL].get())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(memVal > aVal)
	env.setHalfCarryFlag(isHalfBorrow(aVal, memVal))

	subVal := aVal - memVal

	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CP (%v)\n", regHL)
	}
	return 8
}

// cp8BitImm compares register A with an immediate value and sets flags
// accordingly. The semantics are the same as the SUB operator, but the result
// value is not saved.
func cp8BitImm(env *environment) int {
	aVal := env.regs8[regA].get()
	imm := env.incrementPC()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(aVal < imm)
	env.setHalfCarryFlag(isHalfBorrow(aVal, imm))

	subVal := aVal - imm

	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

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
func daa(env *environment) int {
	aVal := env.regs8[regA].get()

	var correction uint8

	// Check if there was a half borrow or if the least significant nibble
	// (which represents the first decimal digit) is overfilled. If it is, a
	// correction needs to be performed.
	if env.getHalfCarryFlag() || (!env.getSubtractFlag() && (aVal&0xF) > 9) {
		correction = 0x06
	}

	// Check if there was a full borrow or if the most significant nibble
	// (which represents the second decimal digit) is overfilled (I think?). If
	// it is, a correction needs to be performed.
	if env.getCarryFlag() || (!env.getSubtractFlag() && aVal > 0x99) {
		correction |= 0x60
		env.setCarryFlag(true)
	} else {
		env.setCarryFlag(false)
	}

	// The direction of the correction depends on what the last operation was
	if env.getSubtractFlag() {
		aVal = env.regs8[regA].set(aVal - correction)
	} else {
		aVal = env.regs8[regA].set(aVal + correction)
	}

	env.setZeroFlag(aVal == 0)
	env.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("DAA\n")
	}

	return 4
}
