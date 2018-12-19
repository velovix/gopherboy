package main

import "fmt"

// ld loads the value of reg2 into reg1.
func ld(state *State, reg1, reg2 registerType) int {
	state.regs8[reg1].set(state.regs8[reg2].get())

	if printInstructions {
		fmt.Printf("LD %v,%v (%#x,%#x)\n",
			reg1, reg2,
			state.regs8[reg1].get(), state.regs8[reg2].get())
	}
	return 4
}

// ldHLToSP puts the value of register HL into register SP.
func ldHLToSP(state *State) int {
	hlVal := state.regs16[regHL].get()

	state.regs16[regSP].set(hlVal)

	if printInstructions {
		fmt.Printf("LD %v,%v\n", regSP, regHL)
	}
	return 8
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(state *State, reg1, reg2 registerType) int {
	state.mmu.set(state.regs16[reg1].get(), state.regs8[reg2].get())

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", reg1, reg2)
	}
	return 12
}

// ldFromMem loads the value in the memory address specified by reg2 into reg1.
func ldFromMem(state *State, reg1, reg2 registerType) int {
	val := state.mmu.at(state.regs16[reg2].get())
	state.regs8[reg1].set(val)

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", reg1, reg2)
	}

	return 8
}

// ld8BitImm loads an 8-bit immediate value into the given register.
func ld8BitImm(state *State, reg registerType) int {
	imm := state.incrementPC()

	state.regs8[reg].set(imm)

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 8
}

// ld16BitImm loads a 16-bit immediate value into the given 16-bit register.
func ld16BitImm(state *State, reg registerType) int {
	imm := combine16(state.incrementPC(), state.incrementPC())
	state.regs16[reg].set(imm)

	if printInstructions {
		fmt.Printf("LD %v,%#x\n", reg, imm)
	}
	return 12
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())
	aVal := state.regs8[regA].get()

	state.mmu.set(imm, aVal)

	if printInstructions {
		fmt.Printf("LD (%#x),%v\n", imm, regA)
	}
	return 16
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())
	memVal := state.mmu.at(imm)

	state.regs8[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%#x)\n", regA, imm)
	}
	return 16
}

// ld8BitImmToMemHL loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToMemHL(state *State) int {
	imm := state.incrementPC()

	state.mmu.set(state.regs16[regHL].get(), imm)

	if printInstructions {
		fmt.Printf("LD (%v),%#x\n", regHL, imm)
	}
	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(state.regs16[regSP].get())
	state.mmu.set(imm, lower)
	state.mmu.set(imm+1, upper)

	if printInstructions {
		fmt.Printf("LD (%#x),%v\n", imm, regSP)
	}
	return 20
}

// ldToMemC saves the value of register A at the memory address
// 0xFF00+register C.
func ldToMemC(state *State) int {
	aVal := state.regs8[regA].get()
	addr := uint16(state.regs8[regC].get()) + 0xFF00

	state.mmu.set(addr, aVal)

	if printInstructions {
		fmt.Printf("LD (%v),%v\n", regC, regA)
	}
	return 8
}

// ldFromMemC loads the value at memory address 0xFF00 + register C into
// register A.
func ldFromMemC(state *State) int {
	addr := uint16(state.regs8[regC].get()) + 0xFF00
	memVal := state.mmu.at(addr)

	state.regs8[regA].set(memVal)

	if printInstructions {
		fmt.Printf("LD %v,(%v)\n", regA, regC)
	}
	return 8
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(state *State) int {
	state.mmu.set(state.regs16[regHL].get(), state.regs8[regA].get())

	state.regs16[regHL].set(state.regs16[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD (%v+),%v\n", regHL, regA)
	}
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(state *State) int {
	state.mmu.set(state.regs16[regHL].get(), state.regs8[regA].get())

	state.regs16[regHL].set(state.regs16[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD (%v-),%v\n", regHL, regA)
	}
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(state *State) int {
	memVal := state.mmu.at(state.regs16[regHL].get())
	state.regs8[regA].set(memVal)

	state.regs16[regHL].set(state.regs16[regHL].get() + 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v+)\n", regA, regHL)
	}
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(state *State) int {
	memVal := state.mmu.at(state.regs16[regHL].get())
	state.regs8[regA].set(memVal)

	state.regs16[regHL].set(state.regs16[regHL].get() - 1)

	if printInstructions {
		fmt.Printf("LD %v,(%v-)\n", regA, regHL)
	}
	return 8
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(state *State) int {
	offset := state.incrementPC()

	state.mmu.set(0xFF00+uint16(offset), state.regs8[regA].get())

	if printInstructions {
		fmt.Printf("LDH (%#x),%v (%#x)\n", offset, regA, 0xFF00+uint16(offset))
	}
	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(state *State) int {
	offset := state.incrementPC()

	fromMem := state.mmu.at(0xFF00 + uint16(offset))
	state.regs8[regA].set(fromMem)

	if printInstructions {
		fmt.Printf("LDH %v,(%#x)\n", regA, offset)
	}
	return 12
}

// ldhl loads an 8-bit immediate value and adds it to the stack pointer. The
// result is stored in register HL.
func ldhl(state *State) int {
	immUnsigned := state.incrementPC()
	imm := int8(immUnsigned)
	spVal := state.regs16[regSP].get()

	// This instruction's behavior for the carry and half carry flags is very
	// weird.
	//
	// When checking for a carry and half carry, the immediate value is treated
	// as _unsigned_ for some reason and only the lowest 8 bits of the stack
	// pointer are considered. HL does not play into this calculation at all.
	lowerSP, _ := split16(spVal)
	state.setHalfCarryFlag(isHalfCarry(lowerSP, immUnsigned))
	state.setCarryFlag(isCarry(lowerSP, immUnsigned))

	state.regs16[regHL].set(uint16(int(imm) + int(spVal)))

	state.setZeroFlag(false)
	state.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("LD %v,%v+%#x\n", regHL, regSP, imm)
	}
	return 12
}

// push decrements the stack pointer by 2, then puts the value of the given
// register at its position.
func push(state *State, reg registerType) int {
	state.pushToStack16(state.regs16[reg].get())

	if printInstructions {
		fmt.Printf("PUSH %v\n", reg)
	}
	return 16
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func pop(state *State, reg registerType) int {
	state.regs16[reg].set(state.popFromStack16())

	if printInstructions {
		fmt.Printf("POP %v\n", reg)
	}
	return 12
}
