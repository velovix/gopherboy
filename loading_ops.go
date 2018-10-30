package main

import "fmt"

// ld loads the value of reg2 into reg1.
func ld(env *environment, reg1, reg2 registerType) int {
	env.regs[reg1].set(env.regs[reg2].get())

	if printInstructions {
		fmt.Printf("LD %v,%v\n", reg1, reg2)
	}
	return 4
}

// ldHLToSP puts the value of register HL into register SP.
func ldHLToSP(env *environment) int {
	hlVal := env.regs[regHL].get()

	env.regs[regSP].set(hlVal)

	if printInstructions {
		fmt.Printf("LD %v,%v\n", regSP, regHL)
	}
	return 8
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(env *environment, reg1, reg2 registerType) int {
	env.mmu.set(env.regs[reg1].get(), uint8(env.regs[reg2].get()))

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", reg1, reg2)
	}
	return 12
}

// ldFromMem loads the value in the memory address specified by reg2 into reg1.
func ldFromMem(env *environment, reg1, reg2 registerType) int {
	val := env.mmu.at(env.regs[reg2].get())
	env.regs[reg1].set(uint16(val))

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", reg1, reg2)
	}

	return 8
}

// ld8BitImm loads an 8-bit immediate value into the given register.
func ld8BitImm(env *environment, reg registerType) int {
	imm := env.incrementPC()

	env.regs[reg].set(uint16(imm))

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 8
}

// ld16BitImm loads a 16-bit immediate value into the given 16-bit register.
func ld16BitImm(env *environment, reg registerType) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	env.regs[reg].set(imm)

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 12
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	aVal := uint8(env.regs[regA].get())

	env.mmu.set(imm, aVal)

	if printInstructions {
		fmt.Printf("LD (%#x),%v\n", imm, regA)
	}
	return 16
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	memVal := uint16(env.mmu.at(imm))

	env.regs[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%#x)\n", regA, imm)
	}
	return 16
}

// ld8BitImmToMemHL loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToMemHL(env *environment) int {
	imm := env.incrementPC()

	env.mmu.set(env.regs[regHL].get(), imm)

	if printInstructions {
		fmt.Printf("LD (%v),%#x\n", regHL, imm)
	}
	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(env.regs[regSP].get())
	env.mmu.set(imm, lower)
	env.mmu.set(imm+1, upper)

	if printInstructions {
		fmt.Printf("LD (%#x),%v\n", imm, regSP)
	}
	return 20
}

// ldToMemC saves the value of register A at the memory address
// 0xFF00+register C.
func ldToMemC(env *environment) int {
	aVal := uint8(env.regs[regA].get())
	addr := env.regs[regC].get() + 0xFF00

	env.mmu.set(addr, aVal)

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", regC, regA)
	}
	return 8
}

// ldFromMemC loads the value at memory address 0xFF00 + register C into
// register A.
func ldFromMemC(env *environment) int {
	addr := env.regs[regC].get() + 0xFF00
	memVal := uint16(env.mmu.at(addr))

	env.regs[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", regA, regC)
	}
	return 8
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(env *environment) int {
	env.mmu.set(env.regs[regHL].get(), uint8(env.regs[regA].get()))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD (%v+),%v\n", regHL, regA)
	}
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(env *environment) int {
	env.mmu.set(env.regs[regHL].get(), uint8(env.regs[regA].get()))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD (%v-),%v\n", regHL, regA)
	}
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(env *environment) int {
	memVal := env.mmu.at(env.regs[regHL].get())
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v+)\n", regA, regHL)
	}
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(env *environment) int {
	memVal := env.mmu.at(env.regs[regHL].get())
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v-)\n", regA, regHL)
	}
	return 8
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(env *environment) int {
	offset := env.incrementPC()

	env.mmu.set(0xFF00+uint16(offset), uint8(env.regs[regA].get()))

	if printInstructions {
		fmt.Printf("LDH (%#x),%v\n", offset, regA)
	}
	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(env *environment) int {
	offset := env.incrementPC()

	fromMem := env.mmu.at(0xFF00 + uint16(offset))
	env.regs[regA].set(uint16(fromMem))

	if printInstructions {
		fmt.Printf("LDH %v,(%#x)\n", regA, offset)
	}
	return 12
}

// ldhl loads an 8-bit immediate value and adds it to the stack pointer. The
// result is stored in register HL.
func ldhl(env *environment) int {
	imm := uint16(env.incrementPC())
	spVal := env.regs[regPC].get()

	env.setHalfCarryFlag(isHalfCarry16(spVal, imm))
	env.setCarryFlag(isCarry16(spVal, imm))

	env.regs[regHL].set(imm + spVal)

	env.setZeroFlag(false)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("LD %v,%v+%#x\n", regHL, regSP, imm)
	}
	return 12
}

// push decrements the stack pointer by 2, then puts the value of the given
// register at its position.
func push(env *environment, reg registerType) int {
	env.pushToStack16(env.regs[reg].get())

	if printInstructions {
		fmt.Printf("PUSH %v\n", reg)
	}
	return 16
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func pop(env *environment, reg registerType) int {
	env.regs[reg].set(env.popFromStack16())

	if printInstructions {
		fmt.Printf("POP %v\n", reg)
	}
	return 12
}
