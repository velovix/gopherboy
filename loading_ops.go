package main

// ld loads the value of reg2 into reg1.
func ld(env *environment, reg1, reg2 registerType) int {
	env.regs[reg1].set(env.regs[reg2].get())

	//fmt.Printf("LD %v,%v\n", reg1, reg2)
	return 4
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(env *environment, reg1, reg2 registerType) int {
	env.mbc.set(env.regs[reg1].get(), uint8(env.regs[reg2].get()))

	//fmt.Printf("LD (%v),%v\n", reg1, reg2)
	return 12
}

// ldFromMem loads the value in the memory address specified by reg2 into reg1.
func ldFromMem(env *environment, reg1, reg2 registerType) int {
	val := env.mbc.at(env.regs[reg2].get())
	env.regs[reg1].set(uint16(val))

	//fmt.Printf("LD %v,(%v)\n", reg1, reg2)

	return 8
}

// ld8BitImm loads an 8-bit immediate value into the given register.
func ld8BitImm(env *environment, reg registerType) int {
	imm := env.incrementPC()

	env.regs[reg].set(uint16(imm))

	//fmt.Printf("LD %v,%#x\n", reg, imm)
	return 8
}

// ld16BitImm loads a 16-bit immediate value into the given 16-bit register.
func ld16BitImm(env *environment, reg registerType) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	env.regs[reg].set(imm)

	//fmt.Printf("LD %v,%#x\n", reg, imm)
	return 12
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	aVal := uint8(env.regs[regA].get())

	env.mbc.set(imm, aVal)

	return 16
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	memVal := uint16(env.mbc.at(imm))

	env.regs[regA].set(memVal)

	return 16
}

// ld8BitImmToHLMem loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToHLMem(env *environment) int {
	imm := env.incrementPC()

	env.mbc.set(env.regs[regHL].get(), imm)

	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(env.regs[regSP].get())
	env.mbc.set(imm, lower)
	env.mbc.set(imm+1, upper)

	//fmt.Printf("LD (%#x),%v\n", imm, regSP)
	return 20
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(env *environment) int {
	env.mbc.set(env.regs[regHL].get(), uint8(env.regs[regA].get()))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	//fmt.Printf("LD (%v+),%v\n", regHL, regA)
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(env *environment) int {
	env.mbc.set(env.regs[regHL].get(), uint8(env.regs[regA].get()))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	//fmt.Printf("LD (%v-),%v\n", regHL, regA)
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(env *environment) int {
	memVal := env.mbc.at(env.regs[regHL].get())
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	//fmt.Printf("LD %v,(%v+)\n", regA, regHL)
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(env *environment) int {
	memVal := env.mbc.at(env.regs[regHL].get())
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	//fmt.Printf("LD %v,(%v-)\n", regA, regHL)
	return 8
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(env *environment) int {
	offset := env.incrementPC()

	env.mbc.set(0xFF00+uint16(offset), uint8(env.regs[regA].get()))

	//fmt.Printf("LDH %#x,%v\n", offset, regA)
	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(env *environment) int {
	offset := env.incrementPC()

	fromMem := env.mbc.at(0xFF00 + uint16(offset))
	env.regs[regA].set(uint16(fromMem))

	//fmt.Printf("LDH %v,%#x\n", regA, offset)
	return 12
}

// push decrements the stack pointer by 2, then puts the value of the given
// register at its position.
func push(env *environment, reg registerType) int {
	env.pushToStack16(env.regs[reg].get())

	//fmt.Printf("PUSH %v\n", reg)
	return 16
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func pop(env *environment, reg registerType) int {
	env.regs[reg].set(env.popFromStack16())

	//fmt.Printf("POP %v\n", reg)
	return 12
}
