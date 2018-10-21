package main

import (
	"fmt"
	"math/bits"
)

// TODO(velovix): Document what instructions do to flags

func nop(env *environment) int {
	fmt.Printf("NOP\n")
	return 4
}

// ld loads the value of reg2 into reg1.
func ld(env *environment, reg1, reg2 registerType) int {
	env.regs[reg1].set(env.regs[reg2].get())

	fmt.Printf("LD %v,%v\n", reg1, reg2)
	return 4
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(env *environment, reg1, reg2 registerType) int {
	env.mem[env.regs[reg1].get()] = uint8(env.regs[reg2].get())

	fmt.Printf("LD %v,%v\n", reg2, reg1)
	return 12
}

// ld8BitImm loads an 8-bit immediate value into the given register.
func (env *environment) ld8BitImm(reg registerType) int {
	imm := env.incrementPC()

	env.regs[reg].set(uint16(imm))

	fmt.Printf("LD %v,%#x\n", reg, imm)
	return 8
}

// ld16BitImm loads an 16-bit immediate value into the given register.
func (env *environment) ld16BitImm(reg registerType) int {
	lower := env.incrementPC()
	upper := env.incrementPC()

	imm := combine(lower, upper)
	env.regs[reg].set(imm)

	fmt.Printf("LD %v,%#x\n", reg, imm)
	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func (env *environment) ldSPToMem() int {
	imm := combine(env.incrementPC(), env.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(env.regs[regSP].get())
	env.mem[imm] = lower
	env.mem[imm+1] = upper

	fmt.Printf("LD %#x,%v\n", imm, regSP)
	return 20
}

// addToA adds the value of reg into register A.
func (env *environment) addToA(reg registerType) int {
	aVal := env.regs[regA].get()
	regVal := env.regs[reg].get()

	env.setHalfCarryFlag(isHalfCarry(uint8(aVal), uint8(regVal)))
	env.setCarryFlag(isCarry(uint8(aVal), uint8(regVal)))

	aVal = env.regs[regA].set(aVal + regVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)

	fmt.Printf("ADD A,%v\n", reg)
	if reg == regHL {
		// It takes longer to add from a 16-bit register
		return 8
	}
	return 4
}

// addToHL adds the value of reg into register HL.
func (env *environment) addToHL(reg registerType) int {
	hlVal := env.regs[regHL].get()
	regVal := env.regs[reg].get()

	env.setHalfCarryFlag(isHalfCarry16(hlVal, regVal))
	env.setCarryFlag(isCarry16(hlVal, regVal))

	hlVal = env.regs[regHL].set(hlVal + regVal)

	env.setSubtractFlag(false)

	fmt.Printf("ADD HL,%v\n", reg)
	return 8
}

// addToSP loads an immediate 8-bit value and adds it to the stack pointer
// register.
func (env *environment) addToSP() int {
	imm := asSigned(env.incrementPC())

	// TODO(velovix): Set flags

	env.regs[regSP].set(uint16(int(env.regs[regSP].get()) + int(imm)))

	fmt.Printf("ADD SP,%#x\n", imm)
	return 16
}

// sub subtracts the value of reg from register A.
func (env *environment) sub(reg registerType) int {
	// TODO(velovix): Set flags

	env.regs[regA].set(env.regs[regA].get() - env.regs[reg].get())

	fmt.Printf("SUB %v,%v\n", regA, reg)
	if reg == regHL {
		// 8-bit HL operations take longer
		return 8
	}
	return 4
}

// dec decrements the value of reg.
func (env *environment) dec(reg registerType) int {
	_, upperNibbleBefore := split(uint8(env.regs[reg].get()))

	env.regs[reg].set(env.regs[reg].get() - 1)

	_, upperNibbleAfter := split(uint8(env.regs[reg].get()))

	// Only the 8-bit operation uses flags
	if env.regs[reg].size() == 8 {
		// A half carry occurs if the upper nibble has changed at all
		env.setHalfCarryFlag(upperNibbleBefore == upperNibbleAfter)
		env.setZeroFlag(env.regs[reg].get() == 0)
		env.setSubtractFlag(true)
	}

	fmt.Printf("DEC %v\n", reg)
	if reg == regHL {
		// 8-bit HL operations take longer
		return 12
	}
	return 4
}

// and performs a bitwise & on the given register and register A, storing the
// result in register A.
func (env *environment) and(reg registerType) int {
	env.regs[regA].set(env.regs[regA].get() & env.regs[reg].get())

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	fmt.Printf("AND %v\n", reg)
	if reg == regHL {
		// 8-bit HL operations take longer
		return 8
	}
	return 4
}

// and performs a bitwise & on register A and an immediate value, storing the
// result in register A.
func (env *environment) and8BitImm() int {
	imm := uint16(env.incrementPC())

	env.regs[regA].set(env.regs[regA].get() & imm)

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	fmt.Printf("AND %#x\n", imm)
	return 8
}

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func (env *environment) ldiToMem() int {
	env.mem[env.regs[regHL].get()] = uint8(env.regs[regA].get())

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	fmt.Printf("LD %v+,%v\n", regHL, regA)
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func (env *environment) lddToMem() int {
	env.mem[env.regs[regHL].get()] = uint8(env.regs[regA].get())

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	fmt.Printf("LD %v-,%v\n", regHL, regA)
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func (env *environment) ldiFromMem() int {
	memVal := env.mem[env.regs[regHL].get()]
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	fmt.Printf("LD %v,%v+\n", regA, regHL)
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func (env *environment) lddFromMem() int {
	memVal := env.mem[env.regs[regHL].get()]
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	fmt.Printf("LD %v,%v-\n", regA, regHL)
	return 8
}

// inc increments the given register by 1.
func (env *environment) inc(reg registerType) int {
	originalVal := env.regs[reg].get()

	env.regs[reg].set(originalVal + 1)

	// Only 8 bit register increments have flag effects
	if env.regs[reg].size() == 8 {
		env.setZeroFlag(env.regs[reg].get() == 0)
		env.setSubtractFlag(false)
		// A half carry occurs only if the bottom 4 bits of the number are 1,
		// meaning all those "slots" are "filled"
		env.setHalfCarryFlag(originalVal&0x0F == 0x0F)
	}

	fmt.Printf("INC %v\n", reg)
	if env.regs[reg].size() == 8 {
		return 4
	}
	// Incrementing 16 bit registers takes longer
	return 8
}

// push decrements the stack pointer by 2, then puts the value of the given
// register at its position.
func (env *environment) push(reg registerType) int {
	env.pushToStack16(env.regs[reg].get())

	fmt.Printf("PUSH %v\n", reg)
	return 16
}

// pop loads the two bytes at the top of the stack in the given register and
// increments the stack pointer by 2.
func (env *environment) pop(reg registerType) int {
	upper := env.popFromStack()
	lower := env.popFromStack()

	env.regs[reg].set(combine(lower, upper))

	fmt.Printf("POP %v\n", reg)
	return 12
}

// rst pushes the current program counter to the stack and jumps to the given
// address.
func (env *environment) rst(address uint16) int {
	env.push(regPC)

	env.regs[regPC].set(address)

	return 16
}

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func (env *environment) jr() int {
	offset := asSigned(env.incrementPC())
	// If we're going backwards, skip past the offset portion of the
	// instruction
	if offset < 0 {
		offset--
	}
	env.regs[regPC].set(uint16(int(env.regs[regPC].get()) + int(offset)))

	return 12
}

// jrIfFlag loads an offset value, then jumps to the operation at address PC +
// offset if the given flag is at the expected setting.
func (env *environment) jrIfFlag(flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask
	offset := asSigned(env.incrementPC())

	fmt.Printf("JR %#x==%v,%#x\n", flagMask, isSet, offset)
	if flagState == isSet {
		// If we're going backwards, skip past the offset portion of the
		// instruction
		if offset < 0 {
			offset--
		}
		env.regs[regPC].set(uint16(int(env.regs[regPC].get()) + int(offset)))
		return 12
	} else {
		return 8
	}
}

// jp loads a 16-bit address and jumps to it.
func (env *environment) jp() int {
	address := combine(env.incrementPC(), env.incrementPC())
	env.regs[regPC].set(address)

	return 16
}

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address.
func (env *environment) call() int {
	address := combine(env.incrementPC(), env.incrementPC())
	env.push(regPC)

	env.regs[regPC].set(address)

	fmt.Printf("CALL %#x\n", address)
	return 24
}

// ret pops a 16-bit address from the stack and jumps to it.
func (env *environment) ret() int {
	upper := env.popFromStack()
	lower := env.popFromStack()

	addr := combine(lower, upper)

	env.regs[regPC].set(addr)

	fmt.Printf("RET %#x\n", addr)
	return 16
}

// retIfFlag pops a 16-bit address from the stack and jumps to it, but only if
// the given flag is at the expected value.
func (env *environment) retIfFlag(flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask

	var opClocks int
	if flagState == isSet {
		upper := env.popFromStack()
		lower := env.popFromStack()

		addr := combine(lower, upper)

		env.regs[regPC].set(addr)
		opClocks = 20
	} else {
		opClocks = 8
	}

	fmt.Printf("RET %#x - %v\n", flagMask, isSet)
	return opClocks
}

// reti pops a 16-bit address from the stack and jumps to it, then enables
// interrupts.
func (env *environment) reti() int {
	upper := env.popFromStack()
	lower := env.popFromStack()

	addr := combine(lower, upper)

	env.regs[regPC].set(addr)

	// TODO(velovix): Enable interrupts when we have those
	fmt.Printf("RETI\n")
	return 16
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func (env *environment) ldhToMem() int {
	offset := env.incrementPC()

	env.mem[0xFF00+uint16(offset)] = uint8(env.regs[regA].get())

	fmt.Printf("LDH %#x,%v\n", offset, regA)
	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func (env *environment) ldhFromMem() int {
	offset := env.incrementPC()

	fromMem := env.mem[0xFF00+uint16(offset)]
	env.regs[regA].set(uint16(fromMem))

	fmt.Printf("LDH %v,%#x\n", regA, offset)
	return 12
}

// rlca bit rotates register A left by one, which is equivalent to a left bit
// shift where the most significant bit is carried over to the least
// significant bit. This bit is also stored in the carry flag.
func (env *environment) rlca() int {
	rotated := bits.RotateLeft8(uint8(env.regs[regA].get()), 1)
	env.regs[regA].set(uint16(rotated))

	env.setZeroFlag(env.regs[regA].get() == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	carryBit := env.regs[regA].get() & 0x01
	env.setCarryFlag(carryBit == 1)

	fmt.Printf("RLCA\n")
	return 4
}

// di sets the master interrupt flag to false, disabling all interrupt
// handling.
func di(env *environment) int {
	env.interruptsEnabled = false

	fmt.Printf("DI\n")
	return 4
}

// di sets the master interrupt flag to true. Interrupts may still be disabled
// using the interrupt flags memory register, however.
func ei(env *environment) int {
	env.interruptsEnabled = true

	fmt.Printf("EI\n")
	return 4
}

func halt(env *environment) int {
	env.waitingForInterrupts = true

	fmt.Printf("HALT\n")
	return 4
}

// runOpcode runs the operation that maps to the given opcode.
func runOpcode(env *environment, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	fmt.Printf("%#x: ", opcode)

	switch {
	// NOP
	case opcode == 0x00:
		return nop(env), nil
	// The CB prefix. Comes before an extended instruction set.
	case opcode == 0xCB:
		panic("CB prefix instructions are not supported")
	// STOP 0
	case opcode == 0x10:
		// TODO(velovix): Blank the screen when I eventually have graphics support
		// TODO(velovix): Have the system wake up on button press when I have
		// interrupts implemented

		fmt.Println("STOP called, freezing execution")
		for {
		}

		return 4, nil
	// DI
	case opcode == 0xF3:
		return di(env), nil
	// EI
	case opcode == 0xFB:
		return ei(env), nil
	// HALT
	case opcode == 0x76:
		return halt(env), nil
	// LD BC,A
	case opcode == 0x02:
		return ldToMem(env, regBC, regA), nil
	// LD DE,A
	case opcode == 0x12:
		return ldToMem(env, regDE, regA), nil
	// LD A,BC
	case opcode == 0x0A:
		return ld(env, regA, regBC), nil
	// LD A,DE
	case opcode == 0x1A:
		return ld(env, regA, regDE), nil
	// LDH (a8),A
	case opcode == 0xE0:
		return env.ldhToMem(), nil
	// LDH A,(a8)
	case opcode == 0xF0:
		return env.ldhFromMem(), nil
	// JR NZ,n
	case opcode == 0x20:
		return env.jrIfFlag(zeroFlag, false), nil
	// JR Z,n
	case opcode == 0x28:
		return env.jrIfFlag(zeroFlag, true), nil
	// JR NC,n
	case opcode == 0x30:
		return env.jrIfFlag(halfCarryFlag, false), nil
	// JR C,n
	case opcode == 0x30:
		return env.jrIfFlag(halfCarryFlag, true), nil
	// JP nn
	case opcode == 0xC3:
		return env.jp(), nil
	// RET
	case opcode == 0xC9:
		return env.ret(), nil
	// RET NZ
	case opcode == 0xC0:
		return env.retIfFlag(zeroFlag, false), nil
	// RET Z
	case opcode == 0xC8:
		return env.retIfFlag(zeroFlag, true), nil
	// RET NC
	case opcode == 0xD0:
		return env.retIfFlag(carryFlag, false), nil
	// RET C
	case opcode == 0xD8:
		return env.retIfFlag(carryFlag, true), nil
	// RETI
	case opcode == 0xD9:
		return env.reti(), nil
	// CALL address
	case opcode == 0xCD:
		return env.call(), nil
	// RLCA
	case opcode == 0x07:
		return env.rlca(), nil
	// Opcodes for loading to and from memory with HL
	case opcode == 0x2A:
		return env.ldiFromMem(), nil
	case opcode == 0x3A:
		return env.lddFromMem(), nil
	case opcode == 0x22:
		return env.ldiToMem(), nil
	case opcode == 0x32:
		return env.lddToMem(), nil
	// JR nn
	case opcode == 0x18:
		return env.jr(), nil
	// ADD HL,nn"
	case lowerNibble == 0x09 && upperNibble <= 0x03:
		return env.addToHL(indexTo16BitRegister[upperNibble]), nil
	// LD nn,SP
	case opcode == 0x08:
		return env.ldSPToMem(), nil
	// LD 16BitReg,nn
	case lowerNibble == 0x01 && upperNibble <= 0x03:
		return env.ld16BitImm(indexTo16BitRegister[upperNibble]), nil
	// LD reg1,reg2
	case opcode >= 0x40 && opcode <= 0x7F:
		// TODO(velovix): This behavior just isn't correct. There are a few
		// address-based ops in this block.
		row := upperNibble - 0x04
		col := lowerNibble

		toRegIndex := (row*0x0F + col) / 8
		fromRegIndex := col % 8

		return ld(env, indexToRegister[toRegIndex], indexToRegister[fromRegIndex]), nil
	// LD reg,n
	case upperNibble <= 0x03 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return env.ld8BitImm(indexToRegister[regIndex]), nil
	// ADD A,reg
	case opcode >= 0x80 && opcode <= 0x87:
		return env.addToA(indexToRegister[lowerNibble]), nil
	// DEC reg
	case upperNibble <= 0x03 && (lowerNibble == 0x05 || lowerNibble == 0x0D):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0D {
			regIndex += 1
		}

		return env.dec(indexToRegister[regIndex]), nil
	// AND reg
	case opcode >= 0xA0 && opcode <= 0xA7:
		return env.and(indexToRegister[lowerNibble]), nil
	// AND n
	case opcode == 0xE6:
		return env.and8BitImm(), nil
	// SUB reg
	case opcode >= 0x90 && opcode <= 0x96:
		return env.sub(indexToRegister[lowerNibble]), nil
	// INC reg
	case upperNibble <= 0x03 && (lowerNibble == 0x04 || lowerNibble == 0x0C):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0C {
			regIndex++
		}

		return env.inc(indexToRegister[regIndex]), nil
	// INC 16BitReg
	case upperNibble <= 0x03 && lowerNibble == 0x03:
		return env.inc(indexTo16BitRegister[upperNibble]), nil
	// POP 16BitReg
	case upperNibble >= 0x0C && lowerNibble == 0x01:
		regIndex := upperNibble - 0x0C
		return env.pop(indexTo16BitRegister[regIndex]), nil
	// PUSH 16BitReg
	case upperNibble >= 0x0C && lowerNibble == 0x05:
		regIndex := upperNibble - 0x0C
		return env.push(indexTo16BitRegister[regIndex]), nil
	// RST address
	case upperNibble >= 0x0C && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		restartIndex := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			restartIndex++
		}

		return env.rst(indexToRestartAddress[restartIndex]), nil
	default:
		return 0, fmt.Errorf("unknown opcode %#x", opcode)
	}
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
