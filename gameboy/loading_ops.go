package gameboy

import "fmt"

// makeLD creates an instruction that loads the value of reg2 into reg1.
func makeLD(reg1, reg2 register8) instruction {
	return func(state *State) int {
		reg1.set(reg2.get())

		if printInstructions {
			fmt.Printf("LD %v,%v (%#x,%#x)\n",
				reg1, reg2,
				reg1.get(), reg2.get())
		}
		return 4
	}
}

// ldHLToSP puts the value of register HL into register SP.
func ldHLToSP(state *State) int {
	hlVal := state.regHL.get()

	state.regSP.set(hlVal)

	return 8
}

// makeLDToMem creates an instruction that loads the value of reg2 into the
// memory address specified by reg1.
func makeLDToMem(reg1 register16, reg2 register8) instruction {
	return func(state *State) int {
		state.mmu.set(reg1.get(), reg2.get())

		if printInstructions {
			fmt.Printf("LD (%v),%v\n", reg1, reg2)
		}
		return 8
	}
}

// makeLDFromMem creates an instruction that loads the value in the memory
// address specified by reg2 into reg1.
func makeLDFromMem(reg1 register8, reg2 register16) instruction {
	return func(state *State) int {
		val := state.mmu.at(reg2.get())
		reg1.set(val)

		if printInstructions {
			fmt.Printf("LD %v,(%v)\n", reg1, reg2)
		}

		return 8
	}
}

// makeLD8BitImm creates an instruction that loads an 8-bit immediate value
// into the given register.
func makeLD8BitImm(reg register8) instruction {
	return func(state *State) int {
		imm := state.incrementPC()

		reg.set(imm)

		if printInstructions {
			fmt.Printf("LD %v,%#x\n", reg, imm)
		}
		return 8
	}
}

// makeLD16BitImm creates an instruction that loads a 16-bit immediate value
// into the specified 16-bit register.
func makeLD16BitImm(reg register16) func(*State) int {
	return func(state *State) int {
		imm := combine16(state.incrementPC(), state.incrementPC())
		reg.set(imm)

		if printInstructions {
			fmt.Printf("LD %v,%#x\n", reg, imm)
		}
		return 12
	}
}

// ldTo16BitImmMem saves the value of register A to an address in memory
// specified by a 16-bit immediate value.
func ldTo16BitImmMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())
	aVal := state.regA.get()

	state.mmu.set(imm, aVal)

	return 16
}

// ldFrom16BitImmMem loads the value in memory at the address specified by a
// 16-bit immediate value into register A.
func ldFrom16BitImmMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())
	memVal := state.mmu.at(imm)

	state.regA.set(memVal)

	return 16
}

// makeLD8BitImmToMemHL loads an 8-bit immediate value into the memory address
// specified by the HL register.
func makeLD8BitImmToMemHL(state *State) int {
	imm := state.incrementPC()

	state.mmu.set(state.regHL.get(), imm)

	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(state *State) int {
	imm := combine16(state.incrementPC(), state.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(state.regSP.get())
	state.mmu.set(imm, lower)
	state.mmu.set(imm+1, upper)

	return 20
}

// ldToMemC saves the value of register A at the memory address
// 0xFF00+register C.
func ldToMemC(state *State) int {
	aVal := state.regA.get()
	addr := uint16(state.regC.get()) + 0xFF00

	state.mmu.set(addr, aVal)

	return 8
}

// ldFromMemC loads the value at memory address 0xFF00 + register C into
// register A.
func ldFromMemC(state *State) int {
	addr := uint16(state.regC.get()) + 0xFF00
	memVal := state.mmu.at(addr)

	state.regA.set(memVal)

	return 8
}

// makeLDIToMem loads register A into the memory address specified by register
// HL, then increments register HL.
func makeLDIToMem(state *State) int {
	state.mmu.set(state.regHL.get(), state.regA.get())

	state.regHL.set(state.regHL.get() + 1)

	return 8
}

// makeLDDToMem loads register A into the memory address specified by register
// HL, then decrements register HL.
func makeLDDToMem(state *State) int {
	state.mmu.set(state.regHL.get(), state.regA.get())

	state.regHL.set(state.regHL.get() - 1)

	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(state *State) int {
	memVal := state.mmu.at(state.regHL.get())
	state.regA.set(memVal)

	state.regHL.set(state.regHL.get() + 1)

	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(state *State) int {
	memVal := state.mmu.at(state.regHL.get())
	state.regA.set(memVal)

	state.regHL.set(state.regHL.get() - 1)

	return 8
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(state *State) int {
	offset := state.incrementPC()

	state.mmu.set(0xFF00+uint16(offset), state.regA.get())

	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(state *State) int {
	offset := state.incrementPC()

	fromMem := state.mmu.at(0xFF00 + uint16(offset))
	state.regA.set(fromMem)

	return 12
}

// ldhl loads an 8-bit immediate value and adds it to the stack pointer. The
// result is stored in register HL.
func ldhl(state *State) int {
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

	return 12
}

// makePUSH creates an instruction that decrements the stack pointer by 2, then
// puts the value of the given register at its position.
func makePUSH(reg register16) instruction {
	return func(state *State) int {
		state.pushToStack16(reg.get())

		if printInstructions {
			fmt.Printf("PUSH %v\n", reg)
		}
		return 16
	}
}

// makePOP creates an instruction that loads the two bytes at the top of the
// stack in the given register and increments the stack pointer by 2.
func makePOP(reg register16) instruction {
	return func(state *State) int {
		reg.set(state.popFromStack16())

		if printInstructions {
			fmt.Printf("POP %v\n", reg)
		}
		return 12
	}
}
