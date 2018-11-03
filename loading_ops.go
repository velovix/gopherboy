package main

import "fmt"

// ld loads the value of reg2 into reg1.
func ld(env *environment, reg1, reg2 registerType) int {
	env.regs8[reg1].set(env.regs8[reg2].get())

	if printInstructions {
		fmt.Printf("LD %v,%v (%#x,%#x)\n",
			reg1, reg2,
			env.regs8[reg1].get(), env.regs8[reg2].get())
	}
	return 4
}

// ldHLToSP puts the value of register HL into register SP.
func ldHLToSP(env *environment) int {
	hlVal := env.regs16[regHL].get()

	env.regs16[regSP].set(hlVal)

	if printInstructions {
		fmt.Printf("LD %v,%v\n", regSP, regHL)
	}
	return 8
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(env *environment, reg1, reg2 registerType) int {
	env.mmu.set(env.regs16[reg1].get(), env.regs8[reg2].get())

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", reg1, reg2)
	}
	return 12
}

// ldFromMem loads the value in the memory address specified by reg2 into reg1.
func ldFromMem(env *environment, reg1, reg2 registerType) int {
	val := env.mmu.at(env.regs16[reg2].get())
	env.regs8[reg1].set(val)

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", reg1, reg2)
	}

	return 8
}

// ld8BitImm loads an 8-bit immediate value into the given register.
func ld8BitImm(env *environment, reg registerType) int {
	imm := env.incrementPC()

	env.regs8[reg].set(imm)

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 8
}

// ld16BitImm loads a 16-bit immediate value into the given 16-bit register.
func ld16BitImm(env *environment, reg registerType) int {
	imm := combine16(env.incrementPC(), env.incrementPC())
	env.regs16[reg].set(imm)

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 12
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(env *environment) int {
	imm := combine16(env.incrementPC(), env.incrementPC())
	aVal := env.regs8[regA].get()

	env.mmu.set(imm, aVal)

	if printInstructions {
		fmt.Printf("LD (%#x),%v\n", imm, regA)
	}
	return 16
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(env *environment) int {
	imm := combine16(env.incrementPC(), env.incrementPC())
	memVal := env.mmu.at(imm)

	env.regs8[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%#x)\n", regA, imm)
	}
	return 16
}

// ld8BitImmToMemHL loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToMemHL(env *environment) int {
	imm := env.incrementPC()

	env.mmu.set(env.regs16[regHL].get(), imm)

	if printInstructions {
		fmt.Printf("LD (%v),%#x\n", regHL, imm)
	}
	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(env *environment) int {
	imm := combine16(env.incrementPC(), env.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(env.regs16[regSP].get())
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
	aVal := env.regs8[regA].get()
	addr := uint16(env.regs8[regC].get()) + 0xFF00

	env.mmu.set(addr, aVal)

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", regC, regA)
	}
	return 8
}

// ldFromMemC loads the value at memory address 0xFF00 + register C into
// register A.
func ldFromMemC(env *environment) int {
	addr := uint16(env.regs8[regC].get()) + 0xFF00
	memVal := env.mmu.at(addr)

	env.regs8[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", regA, regC)
	}
	return 8
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(env *environment) int {
	env.mmu.set(env.regs16[regHL].get(), env.regs8[regA].get())

	env.regs16[regHL].set(env.regs16[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD (%v+),%v\n", regHL, regA)
	}
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(env *environment) int {
	env.mmu.set(env.regs16[regHL].get(), env.regs8[regA].get())

	env.regs16[regHL].set(env.regs16[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD (%v-),%v\n", regHL, regA)
	}
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(env *environment) int {
	memVal := env.mmu.at(env.regs16[regHL].get())
	env.regs8[regA].set(memVal)

	env.regs16[regHL].set(env.regs16[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v+)\n", regA, regHL)
	}
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(env *environment) int {
	memVal := env.mmu.at(env.regs16[regHL].get())
	env.regs8[regA].set(memVal)

	env.regs16[regHL].set(env.regs16[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v-)\n", regA, regHL)
	}
	return 8
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(env *environment) int {
	offset := env.incrementPC()

	env.mmu.set(0xFF00+uint16(offset), env.regs8[regA].get())

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
	env.regs8[regA].set(fromMem)

	if printInstructions {
		fmt.Printf("LDH %v,(%#x)\n", regA, offset)
	}
	return 12
}

// ldhl loads an 8-bit immediate value and adds it to the stack pointer. The
// result is stored in register HL.
func ldhl(env *environment) int {
	imm := uint16(env.incrementPC())
	spVal := env.regs16[regSP].get()

	env.setHalfCarryFlag(isHalfCarry16(spVal, imm))
	env.setCarryFlag(isCarry16(spVal, imm))

	env.regs16[regHL].set(imm + spVal)

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
	env.pushToStack16(env.regs16[reg].get())

	if printInstructions {
		fmt.Printf("PUSH %v (%#x)\n", reg, env.regs16[reg].get())
	}
	return 16
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func pop(env *environment, reg registerType) int {
	env.regs16[reg].set(env.popFromStack16())

	if printInstructions {
		fmt.Printf("POP %v (%#x)\n", reg, env.regs16[reg].get())
	}
	return 12
}
