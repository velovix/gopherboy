package main

import (
	"fmt"
)

// TODO(velovix): Document what instructions do to flags

// ld8BitImmToReg loads the given 8-bit immediate value into the given
// register.
func (env *environment) ld8BitImmToReg(reg registerType, imm uint8) {
	env.getReg(reg).set(uint16(imm))
}

// ld16BitImmToReg loads the given 16-bit immediate value into the given
// register.
func (env *environment) ld16BitImmToReg(reg registerType, imm uint16) {
	env.getReg(reg).set(imm)
}

// ldRegToReg loads the value of reg2 into reg1.
func (env *environment) ldRegToReg(reg1, reg2 registerType) {
	env.getReg(reg1).set(env.getReg(reg2).get())
}

// addRegToReg adds the value of reg2 to reg1.
func (env *environment) addRegToReg(reg1, reg2 registerType) {
	a := env.getReg(reg1).get()
	b := env.getReg(reg2).get()

	// Figures out if a half carry has occurred. This algorithm extracts the
	// first four bits of each register, adds them together, and checks the
	// 5th bit to see if it's 1. If it is, that means the addition half-carried.
	// TODO(velovix): Is this the right behavior for 16-bit registers?
	env.setHalfCarryFlag(((a&0xF)+(b&0xF))&0x10 == 0x10)
	// Check the 9th bit to see if an 8-bit addition carried over.
	// TODO(velovix): Is this the right behavior for 16-bit registers?
	env.setCarryFlag((a+b)&0x100 == 0x100)

	env.getReg(reg1).set(env.getReg(reg1).get() + env.getReg(reg2).get())

	env.setZeroFlag(env.getReg(reg1).get() == 0)
	env.setSubtractFlag(false)
}

// subRegToReg subtracts the value of reg2 from reg1.
func (env *environment) subRegToReg(reg1, reg2 registerType) {
	// TODO(velovix): Set flags

	env.getReg(reg1).set(env.getReg(reg1).get() - env.getReg(reg2).get())
}

// decReg decrements the value of reg.
func (env *environment) decReg(reg registerType) {
	topNibbleBefore := uint8(env.getReg(reg).get() >> 4)

	env.getReg(reg).set(env.getReg(reg).get() - 1)

	// A half carry occurs if the top nibble has changed at all
	env.setHalfCarryFlag(topNibbleBefore == uint8(env.getReg(reg).get()>>4))
	env.setZeroFlag(env.getReg(reg).get() == 0)
	env.setSubtractFlag(true)
}

// andRegToReg performs a bitwise & on reg1 and reg2, storing the result in
// reg1.
func (env *environment) andRegToReg(reg1, reg2 registerType) {
	env.getReg(reg1).set(env.getReg(reg1).get() & env.getReg(reg2).get())

	env.setZeroFlag(env.getReg(reg1).get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)
}

// ldiToMem loads reg2 into the memory address specified by reg1, then
// increments reg1.
func (env *environment) ldiToMem(reg1, reg2 registerType) {
	env.mem[env.getReg(reg1).get()] = uint8(env.getReg(reg2).get())

	env.getReg(reg1).set(env.getReg(reg1).get() + 1)
}

// lddToMem loads reg2 into the memory address specified by reg1, then
// decrements reg1.
func (env *environment) lddToMem(reg1, reg2 registerType) {
	env.mem[env.getReg(reg1).get()] = uint8(env.getReg(reg2).get())

	env.getReg(reg1).set(env.getReg(reg1).get() - 1)
}

// ldiFromMem puts the value stored in the memory address specified by reg2
// into reg1, then increments reg2.
func (env *environment) ldiFromMem(reg1, reg2 registerType) {
	memVal := env.mem[env.getReg(reg2).get()]
	env.getReg(reg1).set(uint16(memVal))

	env.getReg(reg2).set(env.getReg(reg2).get() + 1)
}

// lddFromMem puts the value stored in the memory address specified by reg2
// into reg1, then decrements reg2.
func (env *environment) lddFromMem(reg1, reg2 registerType) {
	memVal := env.mem[env.getReg(reg2).get()]
	env.getReg(reg1).set(uint16(memVal))

	env.getReg(reg2).set(env.getReg(reg2).get() - 1)
}

// inc increments the given register by 1.
func (env *environment) inc(reg registerType) {
	originalVal := env.getReg(reg).get()

	env.getReg(reg).set(originalVal + 1)

	// Only 8 bit register increments have flag effects
	if env.getReg(reg).size() == 8 {
		env.setZeroFlag(env.getReg(reg).get() == 0)
		env.setSubtractFlag(false)
		// A half carry occurs only if the bottom 4 bits of the number are 1,
		// meaning all those "slots" are "filled"
		env.setHalfCarryFlag(originalVal&0x0F == 0x0F)
	}
}

// push decrements the stack pointer by 2, then puts the value of the given
// register at its position.
func (env *environment) push(reg registerType) {
	env.getReg(regSP).set(env.getReg(regSP).get() - 2)

	// Put each byte of the register into memory
	env.mem[env.getReg(regSP).get()+1] = uint8(env.getReg(reg).get() >> 4)
	env.mem[env.getReg(regSP).get()] = uint8(env.getReg(reg).get() & 0x0F)
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func (env *environment) pop(reg registerType) {
	upperVal := env.mem[env.getReg(regSP).get()+1]
	lowerVal := env.mem[env.getReg(regSP).get()]

	env.getReg(reg).set(uint16(upperVal<<8) | uint16(lowerVal))

	env.getReg(regSP).set(env.getReg(regSP).get() + 2)
}

// rst pushes the current program counter to the stack and jumps to the given
// address.
func (env *environment) rst(address uint16) {
	env.push(regPC)

	env.getReg(regPC).set(address)
}

// jr jumps to the operation at address PC + offset. In other words, it's a
// jump relative to the current position.
func (env *environment) jr(offset int8) {
	// If we're going backwards, skip past the offset portion of the
	// instruction
	if offset < 0 {
		offset--
	}
	env.getReg(regPC).set(uint16(int(env.getReg(regPC).get()) + int(offset)))
}

// jrIfFlag jumps to the operation at address PC + offset if the given flag is
// at the expected setting.
func (env *environment) jrIfFlag(offset int8, flagMask uint16, isSet bool) {
	flagState := env.getReg(regF).get()&flagMask == flagMask
	if flagState == isSet {
		env.jr(offset)
	}
}

// jp jumps to the given address by setting the program counter.
func (env *environment) jp(address uint16) {
	env.getReg(regPC).set(address)
}

// ret pops a 16-bit address from the stack and jumps to it.
func (env *environment) ret() {
	upperVal := env.mem[env.getReg(regSP).get()+1]
	lowerVal := env.mem[env.getReg(regSP).get()]

	env.getReg(regPC).set(uint16(upperVal<<8) | uint16(lowerVal))

	env.getReg(regSP).set(env.getReg(regSP).get() + 2)
}

// ldhToMem saves the value of register A into the memory address 0xFF00 +
// offset.
func (env *environment) ldhToMem(offset uint8) {
	env.mem[env.getReg(regPC).get()+uint16(offset)] = uint8(env.getReg(regA).get())
}

// runInstruction runs the instruction pointed to by the program counter.
func (env *environment) runInstruction() error {
	opcode := env.incrementPC()

	upperNibble := uint8(opcode >> 4)
	lowerNibble := uint8(opcode & 0x0F)

	switch {
	// NOP
	case opcode == 0x00:
		fmt.Println("NOP")
	// STOP 0
	case opcode == 0x10:
		// TODO(velovix): Blank the screen when I eventually have graphics support
		// TODO(velovix): Have the system wake up on button press when I have
		// interrupts implemented

		fmt.Println("STOP called, freezing execution")
		for {
		}
	case opcode == 0xF3:
		// TODO(velovix): Implement this when I have interrupts
		fmt.Println("Ignoring DI")
	// HALT 1,4
	case opcode == 0x76:
		panic("HALT 1 4 not implemented!")
	// LD BC,A
	case opcode == 0x02:
		env.ldRegToReg(regBC, regA)
		fmt.Printf("LD %v,%v\n", regBC, regA)
	// LD DE,A
	case opcode == 0x12:
		env.ldRegToReg(regDE, regA)
		fmt.Printf("LD %v,%v\n", regDE, regA)
	// LD A,BC
	case opcode == 0x0A:
		env.ldRegToReg(regA, regBC)
		fmt.Printf("LD %v,%v\n", regA, regBC)
	// LD A,DE
	case opcode == 0x1A:
		env.ldRegToReg(regA, regDE)
		fmt.Printf("LD %v,%v\n", regA, regDE)
	// LDH (a8),A
	case opcode == 0xE0:
		offset := env.incrementPC()

		env.ldhToMem(offset)
		fmt.Printf("LDH %#x,%v\n", offset, regA)
	// JR NZ,n
	case opcode == 0x20:
		offset := asSigned(env.incrementPC())

		env.jrIfFlag(offset, zeroFlag, false)
		fmt.Printf("JR NZ,%#x\n", offset)
	// JR Z,n
	case opcode == 0x28:
		offset := asSigned(env.incrementPC())

		env.jrIfFlag(offset, zeroFlag, true)
		fmt.Printf("JR Z,%#x\n", offset)
	// JR NC,n
	case opcode == 0x30:
		offset := asSigned(env.incrementPC())

		env.jrIfFlag(offset, halfCarryFlag, false)
		fmt.Printf("JR NC,%#x\n", offset)
	// JR C,n
	case opcode == 0x30:
		offset := asSigned(env.incrementPC())

		env.jrIfFlag(offset, halfCarryFlag, true)
		fmt.Printf("JR C,%#x\n", offset)
	// JP nn
	case opcode == 0xC3:
		// Load the 16-bit address
		addr := combine(env.incrementPC(), env.incrementPC())

		env.jp(addr)
		fmt.Printf("JP %#x\n", addr)
	// RET
	case opcode == 0xC9:
		env.ret()
		fmt.Printf("RET\n")
	// Opcodes for loading to and from memory with HL
	case opcode == 0x2A:
		env.ldiFromMem(regA, regHL)
		fmt.Printf("LDI %v,%v\n", regA, regHL)
	case opcode == 0x3A:
		env.lddFromMem(regA, regHL)
		fmt.Printf("LDD %v,%v\n", regA, regHL)
	case opcode == 0x22:
		env.ldiToMem(regHL, regA)
		fmt.Printf("LDI %v,%v\n", regHL, regA)
	case opcode == 0x32:
		env.lddToMem(regHL, regA)
		fmt.Printf("LDD %v,%v\n", regHL, regA)
	// JR nn
	case opcode == 0x18:
		offset := asSigned(env.incrementPC())
		env.jr(offset)
		fmt.Printf("JR %#x\n", offset)
	// Check if the opcode is in the "ADD HL,16BitReg1" section
	case lowerNibble == 0x09 && upperNibble <= 0x03:
		env.addRegToReg(regHL, indexTo16BitRegister[upperNibble])
		fmt.Printf("ADD %v,%v\n",
			regHL,
			indexTo16BitRegister[upperNibble])
	// Check if the opcode is in the "LD 16BitReg1,16BitImm" section
	case lowerNibble == 0x01 && upperNibble <= 0x03:
		// Load the 16-bit immediate value
		imm := combine(env.incrementPC(), env.incrementPC())

		env.ld16BitImmToReg(indexTo16BitRegister[upperNibble], imm)
		fmt.Printf("LD %v,%#x\n",
			indexTo16BitRegister[upperNibble],
			imm)
	// LD SP,address - Saves the current stack pointer into memory at the
	// given address
	case opcode == 0x08:
		// Load the 16-bit immediate value
		imm := combine(env.incrementPC(), env.incrementPC())

		// Save each byte of the stack pointer into memory
		env.mem[imm] = uint8(env.getReg(regSP).get() >> 4)
		env.mem[imm+1] = uint8(env.getReg(regSP).get() & 0x0F)

		fmt.Printf("LD %#x,%v\n",
			imm,
			regSP)
	// Check if the opcode is in the "LD reg1,reg2" block
	case opcode >= 0x40 && opcode <= 0x7F:
		row := upperNibble - 0x04
		col := lowerNibble

		toRegIndex := (row*0x0F + col) / 8
		fromRegIndex := col % 8

		env.ldRegToReg(
			indexToRegister[toRegIndex],
			indexToRegister[fromRegIndex])
		fmt.Printf("LD %v,%v\n",
			indexToRegister[toRegIndex],
			indexToRegister[fromRegIndex])
	// Check if the opcode is in one of the "LD reg1,8-bit value" sections
	case upperNibble <= 0x03 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		imm := env.incrementPC()

		env.ld8BitImmToReg(indexToRegister[regIndex], imm)
		fmt.Printf("LD %v,%#x\n",
			indexToRegister[regIndex],
			imm)
	// Check if the opcode is within the "LD reg1,8-bit immediate" sections
	case upperNibble <= 0x03 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		imm := env.incrementPC()

		env.ld8BitImmToReg(indexToRegister[regIndex], imm)
		fmt.Printf("LD %v,%#x\n",
			indexToRegister[regIndex],
			imm)
	// Check if the opcode is in the "ADD A,reg1" section
	case opcode >= 0x80 && opcode <= 0x87:
		env.addRegToReg(regA, indexToRegister[lowerNibble])
		fmt.Printf("ADD %v,%v\n",
			regA,
			indexToRegister[lowerNibble])
	// Check if the opcode is in the "DEC reg1" section
	case upperNibble <= 0x03 && (lowerNibble == 0x05 || lowerNibble == 0x0D):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0D {
			regIndex += 1
		}

		env.decReg(indexToRegister[regIndex])
		fmt.Printf("DEC %v\n", indexToRegister[regIndex])
	// Check if the opcode is in the "AND reg1" section
	case opcode >= 0xA0 && opcode <= 0xA7:
		env.andRegToReg(regA, indexToRegister[lowerNibble])
		fmt.Printf("AND %v\n",
			indexToRegister[lowerNibble])
	// Check if the opcode is in the "SUB reg1" section
	case opcode >= 0x90 && opcode <= 0x96:
		env.subRegToReg(regA, indexToRegister[lowerNibble])
		fmt.Printf("SUB %v\n",
			indexToRegister[lowerNibble])
	// Check if the opcode is in the "INC reg1" sections
	case upperNibble <= 0x03 && (lowerNibble == 0x04 || lowerNibble == 0x0C):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0C {
			regIndex++
		}

		env.inc(indexToRegister[regIndex])
		fmt.Printf("INC %v\n",
			indexToRegister[regIndex])
	// Check if the opcode is in the "INC 16BitReg1" section
	case upperNibble <= 0x03 && lowerNibble == 0x03:
		env.inc(indexTo16BitRegister[upperNibble])
		fmt.Printf("INC %v\n",
			indexTo16BitRegister[upperNibble])
	// Check if the opcode is in the "POP 16BitReg1" section
	case upperNibble >= 0x0C && lowerNibble == 0x01:
		regIndex := upperNibble - 0x0C
		env.pop(indexTo16BitRegister[regIndex])
		fmt.Printf("POP %v\n",
			indexTo16BitRegister[regIndex])
	// Check if the opcode is in the "PUSH 16BitReg1" section
	case upperNibble >= 0x0C && lowerNibble == 0x05:
		regIndex := upperNibble - 0x0C
		env.push(indexTo16BitRegister[regIndex])
		fmt.Printf("PUSH %v\n",
			indexTo16BitRegister[regIndex])
	// Check if the opcode is in the "RST address" section
	case upperNibble >= 0x0C && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		restartIndex := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			restartIndex++
		}

		env.rst(indexToRestartAddress[restartIndex])
		fmt.Printf("RST %#x\n",
			indexToRestartAddress[restartIndex])
	default:
		return fmt.Errorf("unknown opcode %#x", opcode)
	}

	fmt.Printf("SP:%#x | PC:%#x\n", env.getReg(regSP).get(), env.getReg(regPC).get())

	return nil
}

// indexToRegister maps an index value to an 8-bit register. This helps reduce
// repetition for opcode sections that do the same thing on different
// registers, since the Gameboy is consistent about this ordering.
var indexToRegister = map[uint8]registerType{
	0: regB,
	1: regC,
	2: regD,
	3: regE,
	4: regH,
	5: regL,
	6: regHL,
	7: regA,
}

// indexTo16BitRegister maps an index value to a 16-bit register. This serves
// the same purpose as indexToRegister, but for 16-bit registers.
var indexTo16BitRegister = map[uint8]registerType{
	0: regBC,
	1: regDE,
	2: regHL,
	3: regSP,
}

// indexToRestartAddress maps an index value to a place in memory. These values
// are specific to the RST instruction. This helps reduce repetition when
// mapping the different flavors of that instruction from their opcodes.
var indexToRestartAddress = map[uint8]uint16{
	0: 0x0000,
	1: 0x0008,
	2: 0x0010,
	3: 0x0018,
	4: 0x0020,
	5: 0x0028,
	6: 0x0030,
	7: 0x0038,
}
