package main

// add adds the value of reg, an 8-bit register, into register A.
func add(env *environment, reg registerType) int {
	aVal := env.regs[regA].get()
	regVal := env.regs[reg].get()

	env.setHalfCarryFlag(isHalfCarry(uint8(aVal), uint8(regVal)))
	env.setCarryFlag(isCarry(uint8(aVal), uint8(regVal)))

	aVal = env.regs[regA].set(aVal + regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	//fmt.Printf("ADD A,%v\n", reg)
	return 4
}

// addFromMemHL adds the value stored in the memory address specified by HL
// into register A.
func addFromMemHL(env *environment) int {
	aVal := uint8(env.regs[regA].get())
	memVal := env.mbc.at(env.regs[regHL].get())

	env.setHalfCarryFlag(isHalfCarry(aVal, memVal))
	env.setCarryFlag(isCarry(aVal, memVal))

	aVal = uint8(env.regs[regA].set(uint16(aVal + memVal)))

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	//fmt.Printf("ADD A,(HL)\n")
	return 8
}

// addToHL adds the value of reg into register HL.
func addToHL(env *environment, reg registerType) int {
	hlVal := env.regs[regHL].get()
	regVal := env.regs[reg].get()

	env.setHalfCarryFlag(isHalfCarry16(hlVal, regVal))
	env.setCarryFlag(isCarry16(hlVal, regVal))
	env.setSubtractFlag(false)

	hlVal = env.regs[regHL].set(hlVal + regVal)

	//fmt.Printf("ADD HL,%v\n", reg)
	return 8
}

// add8BitImm loads an 8-bit immediate value and adds it to register A, storing
// the results in register A.
func add8BitImm(env *environment) int {
	imm := env.incrementPC()
	aVal := env.regs[regA].get()

	env.setHalfCarryFlag(isHalfCarry(uint8(aVal), imm))
	env.setCarryFlag(isCarry(uint8(aVal), imm))

	aVal = env.regs[regA].set(uint16(imm) + aVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	//fmt.Printf("ADD %v,%#x", regA, imm)
	return 8
}

// addToSP loads an immediate 8-bit value and adds it to the stack pointer
// register.
func addToSP(env *environment) int {
	imm := asSigned(env.incrementPC())

	env.regs[regSP].set(uint16(int(env.regs[regSP].get()) + int(imm)))

	env.setZeroFlag(false)
	env.setSubtractFlag(false)
	// TODO(velovix): Find out what this operation is supposed to do with flags

	//fmt.Printf("ADD SP,%#x\n", imm)
	return 16
}

// sub subtracts the value of reg, an 8-bit register, from register A.
func sub(env *environment, reg registerType) int {
	aVal := env.regs[regA].get()
	regVal := env.regs[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(regVal > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	aVal = env.regs[regA].set(aVal - regVal)

	_, upperNibbleAfter := split(uint8(aVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("SUB %v\n", reg)
	return 4
}

// subFromMemHL subtracts the value in memory at the address specified by HL
// from register A.
func subFromMemHL(env *environment) int {
	aVal := env.regs[regA].get()
	memVal := uint16(env.mbc.at(env.regs[regHL].get()))

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(memVal > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	aVal = env.regs[regA].set(aVal - memVal)

	_, upperNibbleAfter := split(uint8(aVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("SUB (%v)\n", regA, regHL)
	return 8
}

// sub8BitImm loads an 8-bit immediate value and subtracts it from register A,
// storing the result in register A.
func sub8BitImm(env *environment) int {
	imm := uint16(env.incrementPC())
	aVal := env.regs[regA].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(imm > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	aVal = env.regs[regA].set(aVal - imm)

	_, upperNibbleAfter := split(uint8(aVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("SUB %#x\n", imm)
	return 8
}

// and performs a bitwise & on the given register and register A, storing the
// result in register A.
func and(env *environment, reg registerType) int {
	env.regs[regA].set(env.regs[regA].get() & env.regs[reg].get())

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	//fmt.Printf("AND %v\n", reg)
	return 4
}

// andFromMemHL performs a bitwise & on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func andFromMemHL(env *environment) int {
	aVal := env.regs[regA].get()
	memVal := uint16(env.mbc.at(env.regs[regHL].get()))

	aVal = env.regs[regA].set(aVal & memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	//fmt.Printf("AND (%v)\n", regHL)
	return 8
}

// and8BitImm performs a bitwise & on register A and an immediate value,
// storing the result in register A.
func and8BitImm(env *environment) int {
	imm := uint16(env.incrementPC())

	env.regs[regA].set(env.regs[regA].get() & imm)

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	//fmt.Printf("AND %#x\n", imm)
	return 8
}

// or performs a bitwise | on the given register and register A, storing the
// result in register A.
func or(env *environment, reg registerType) int {
	env.regs[regA].set(env.regs[regA].get() | env.regs[reg].get())

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	//fmt.Printf("OR %v\n", reg)
	return 4
}

// orFromMemHL performs a bitwise | on the value in memory at the address
// specified by register HL and register A, storing the result in register A.
func orFromMemHL(env *environment) int {
	aVal := env.regs[regA].get()
	memVal := uint16(env.mbc.at(env.regs[regHL].get()))

	aVal = env.regs[regA].set(aVal | memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	//fmt.Printf("OR (%v)\n", regHL)
	return 8
}

// or8BitImm performs a bitwise | on register A and an immediate value,
// storing the result in register A.
func or8BitImm(env *environment) int {
	imm := uint16(env.incrementPC())

	env.regs[regA].set(env.regs[regA].get() | imm)

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	//fmt.Printf("OR %#x\n", imm)
	return 8
}

// xor performs a bitwise ^ on register A and the given register, storing the
// result in register A.
func xor(env *environment, reg registerType) int {
	aVal := env.regs[regA].get()
	regVal := env.regs[reg].get()

	aVal = env.regs[regA].set(aVal ^ regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	//fmt.Printf("XOR %v\n", reg)
	return 4
}

// xorFromMemHL performs a bitwise ^ on the value in memory specified by
// register HL and register A, storing the result in register A.
func xorFromMemHL(env *environment) int {
	aVal := env.regs[regA].get()
	memVal := uint16(env.mbc.at(env.regs[regHL].get()))

	aVal = env.regs[regA].set(aVal ^ memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	//fmt.Printf("XOR (%v)\n", regHL)
	return 8
}

// inc8Bit increments the given 8-bit register by 1.
func inc8Bit(env *environment, reg registerType) int {
	oldVal := env.regs[reg].get()
	newVal := env.regs[reg].set(oldVal + 1)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	env.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	//fmt.Printf("INC %v\n", reg)
	return 4
}

// inc16Bit increments the given 16-bit register by 1.
func inc16Bit(env *environment, reg registerType) int {
	oldVal := env.regs[reg].get()

	env.regs[reg].set(oldVal + 1)

	//fmt.Printf("INC %v\n", reg)
	return 8
}

// incMemHL increments the value in memory at the address specified by register
// HL.
func incMemHL(env *environment) int {
	addr := env.regs[regHL].get()

	oldVal := env.mbc.at(addr)
	env.mbc.set(addr, env.mbc.at(addr)+1)
	newVal := env.mbc.at(addr)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	// A half carry occurs only if the bottom 4 bits of the number are 1,
	// meaning all those "slots" are "filled"
	env.setHalfCarryFlag(oldVal&0x0F == 0x0F)

	//fmt.Printf("INC (HL)\n")

	return 12
}

// dec8Bit decrements the given 8-bit register by 1.
func dec8Bit(env *environment, reg registerType) int {
	_, upperNibbleBefore := split(uint8(env.regs[reg].get()))

	newVal := env.regs[reg].set(env.regs[reg].get() - 1)

	_, upperNibbleAfter := split(uint8(env.regs[reg].get()))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("DEC %v\n", reg)
	return 4
}

// dec16Bit decrements the given 16-bit register by 1.
func dec16Bit(env *environment, reg registerType) int {
	env.regs[reg].set(env.regs[reg].get() - 1)

	//fmt.Printf("DEC %v\n", reg)
	return 8
}

// decMemHL decrements the value in memory at the address specified by register
// HL.
func decMemHL(env *environment) int {
	addr := env.regs[regHL].get()

	oldVal := env.mbc.at(addr)
	_, upperNibbleBefore := split(oldVal)

	env.mbc.set(addr, env.mbc.at(addr)-1)
	newVal := env.mbc.at(addr)
	_, upperNibbleAfter := split(newVal)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(true)
	// A half borrow occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore == upperNibbleAfter)

	//fmt.Printf("DEC (HL)\n")

	return 12
}

// cp compares the value in register A with the value of the given register and
// sets flags accordingly.  The semantics are the same as the SUB operator, but
// the result value is not saved.
func cp(env *environment, reg registerType) int {
	aVal := env.regs[regA].get()
	regVal := env.regs[reg].get()

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(regVal > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	subVal := aVal - regVal

	_, upperNibbleAfter := split(uint8(subVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("CP %v\n", reg)
	return 4
}

// cpFromMemHL compares the value in register A with the value in memory at the
// address specified by the HL register and sets flags accordingly. The
// semantics are the same as the SUB operator, but the result value is not
// saved.
func cpFromMemHL(env *environment) int {
	aVal := env.regs[regA].get()
	memVal := uint16(env.mbc.at(env.regs[regHL].get()))

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(memVal > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	subVal := aVal - memVal

	_, upperNibbleAfter := split(uint8(subVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("CP (%v)\n" regHL)
	return 8
}

// cp8BitImm compares register A with an immediate value and sets flags
// accordingly. The semantics are the same as the SUB operator, but the result
// value is not saved.
func cp8BitImm(env *environment) int {
	aVal := env.regs[regA].get()
	imm := uint16(env.incrementPC())

	// A carry occurs if the value we're subtracting is greater than register
	// A, meaning that the register A value rolled over
	env.setCarryFlag(imm > aVal)

	_, upperNibbleBefore := split(uint8(aVal))

	subVal := aVal - imm

	_, upperNibbleAfter := split(uint8(subVal))

	// A half carry occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore != upperNibbleAfter)
	env.setZeroFlag(subVal == 0)
	env.setSubtractFlag(true)

	//fmt.Printf("CP %#x\n", imm)
	return 8
}
