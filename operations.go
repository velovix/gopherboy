package main

import (
	"fmt"
	"math/bits"
)

// TODO(velovix): Document what instructions do to flags

// nop does nothing.
func nop(env *environment) int {
	//fmt.Printf("NOP\n")
	return 4
}

// ld loads the value of reg2 into reg1.
func ld(env *environment, reg1, reg2 registerType) int {
	env.regs[reg1].set(env.regs[reg2].get())

	//fmt.Printf("LD %v,%v\n", reg1, reg2)
	return 4
}

// ldToMem loads the value of reg2 into the memory address specified by reg1.
func ldToMem(env *environment, reg1, reg2 registerType) int {
	env.mem[env.regs[reg1].get()] = uint8(env.regs[reg2].get())

	//fmt.Printf("LD (%v),%v\n", reg1, reg2)
	return 12
}

// ldFromMem loads the value in the memory address specified by reg2 into reg1.
func ldFromMem(env *environment, reg1, reg2 registerType) int {
	val := env.mem[env.regs[reg2].get()]
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

// ld16BitImm loads an 16-bit immediate value into the given 16-bit register.
func ld16BitImm(env *environment, reg registerType) int {
	imm := combine(env.incrementPC(), env.incrementPC())
	env.regs[reg].set(imm)

	//fmt.Printf("LD %v,%#x\n", reg, imm)
	return 12
}

// ld8BitImmToHLMem loads an 8-bit immediate value into the memory address
// specified by the HL register.
func ld8BitImmToHLMem(env *environment) int {
	imm := env.incrementPC()

	env.mem[env.regs[regHL].get()] = imm

	return 12
}

// ldSPToMem loads a 16-bit address and saves the stack pointer at that
// address.
func ldSPToMem(env *environment) int {
	imm := combine(env.incrementPC(), env.incrementPC())

	// Save each byte of the stack pointer into memory
	lower, upper := split16(env.regs[regSP].get())
	env.mem[imm] = lower
	env.mem[imm+1] = upper

	//fmt.Printf("LD (%#x),%v\n", imm, regSP)
	return 20
}

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
	memVal := env.mem[env.regs[regHL].get()]

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
	memVal := uint16(env.mem[env.regs[regHL].get()])

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
	memVal := uint16(env.mem[env.regs[regHL].get()])

	aVal = env.regs[regA].set(aVal & memVal)

	env.setZeroFlag(aVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)
	env.setCarryFlag(false)

	//fmt.Printf("AND (%v)\n", regHL)
	return 8
}

// and performs a bitwise & on register A and an immediate value, storing the
// result in register A.
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

// ldiToMem loads register A into the memory address specified by register HL,
// then increments register HL.
func ldiToMem(env *environment) int {
	env.mem[env.regs[regHL].get()] = uint8(env.regs[regA].get())

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	//fmt.Printf("LD (%v+),%v\n", regHL, regA)
	return 8
}

// lddToMem loads register A into the memory address specified by register HL,
// then decrements register HL.
func lddToMem(env *environment) int {
	env.mem[env.regs[regHL].get()] = uint8(env.regs[regA].get())

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	//fmt.Printf("LD (%v-),%v\n", regHL, regA)
	return 8
}

// ldiFromMem puts the value stored in the memory address specified by register
// HL into register A, then increments register HL.
func ldiFromMem(env *environment) int {
	memVal := env.mem[env.regs[regHL].get()]
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() + 1)

	//fmt.Printf("LD %v,(%v+)\n", regA, regHL)
	return 8
}

// lddFromMem puts the value stored in the memory address specified by register
// HL into register A, then decrements register HL.
func lddFromMem(env *environment) int {
	memVal := env.mem[env.regs[regHL].get()]
	env.regs[regA].set(uint16(memVal))

	env.regs[regHL].set(env.regs[regHL].get() - 1)

	//fmt.Printf("LD %v,(%v-)\n", regA, regHL)
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

	oldVal := env.mem[addr]
	env.mem[addr]++
	newVal := env.mem[addr]

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

	oldVal := env.mem[addr]
	_, upperNibbleBefore := split(oldVal)

	env.mem[addr]--
	newVal := env.mem[addr]
	_, upperNibbleAfter := split(newVal)

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(true)
	// A half borrow occurs if the upper nibble has changed at all
	env.setHalfCarryFlag(upperNibbleBefore == upperNibbleAfter)

	//fmt.Printf("DEC (HL)\n")

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

// rst pushes the current program counter to the stack and jumps to the given
// address.
func rst(env *environment, address uint16) int {
	env.pushToStack16(env.regs[regPC].get())

	env.regs[regPC].set(address)

	return 16
}

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func jr(env *environment) int {
	offset := asSigned(env.incrementPC())
	env.relativeJump(int(offset))

	return 12
}

// jrIfFlag loads an offset value, then jumps to the operation at address PC +
// offset if the given flag is at the expected setting.
func jrIfFlag(env *environment, flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask
	offset := asSigned(env.incrementPC())

	//conditional := getConditionalStr(flagMask, isSet)
	//fmt.Printf("JR %v,%#x\n", conditional, offset)
	if flagState == isSet {
		env.relativeJump(int(offset))
		return 12
	} else {
		return 8
	}
}

// jp loads a 16-bit address and jumps to it.
func jp(env *environment) int {
	address := combine(env.incrementPC(), env.incrementPC())
	env.regs[regPC].set(address)

	return 16
}

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address.
func call(env *environment) int {
	address := combine(env.incrementPC(), env.incrementPC())
	env.pushToStack16(env.regs[regPC].get())

	env.regs[regPC].set(address)

	//fmt.Printf("CALL %#x\n", address)
	return 24
}

// ret pops a 16-bit address from the stack and jumps to it.
func ret(env *environment) int {
	addr := env.popFromStack16()
	env.regs[regPC].set(addr)

	//fmt.Printf("RET %#x\n", addr)
	return 16
}

// retIfFlag pops a 16-bit address from the stack and jumps to it, but only if
// the given flag is at the expected value.
func retIfFlag(env *environment, flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask

	var opClocks int
	if flagState == isSet {
		addr := env.popFromStack16()
		env.regs[regPC].set(addr)
		opClocks = 20
	} else {
		opClocks = 8
	}

	//conditional := getConditionalStr(flagMask, isSet)
	//fmt.Printf("RET %v\n", conditional)
	return opClocks
}

// reti pops a 16-bit address from the stack and jumps to it, then enables
// interrupts.
func reti(env *environment) int {
	addr := env.popFromStack16()
	env.regs[regPC].set(addr)

	env.interruptsEnabled = true

	//fmt.Printf("RETI\n")
	return 16
}

// ldhToMem loads an offset value, then saves the value of register A into the
// memory address 0xFF00 + offset.
func ldhToMem(env *environment) int {
	offset := env.incrementPC()

	env.mem[0xFF00+uint16(offset)] = uint8(env.regs[regA].get())

	//fmt.Printf("LDH %#x,%v\n", offset, regA)
	return 12
}

// ldhFromMem loads an offset value, then loads the value at memory address
// 0xFF00 + offset into register A.
func ldhFromMem(env *environment) int {
	offset := env.incrementPC()

	fromMem := env.mem[0xFF00+uint16(offset)]
	env.regs[regA].set(uint16(fromMem))

	//fmt.Printf("LDH %v,%#x\n", regA, offset)
	return 12
}

// rlca bit rotates register A left by one, which is equivalent to a left bit
// shift where the most significant bit is carried over to the least
// significant bit. This bit is also stored in the carry flag.
func rlca(env *environment) int {
	rotated := bits.RotateLeft8(uint8(env.regs[regA].get()), 1)
	env.regs[regA].set(uint16(rotated))

	env.setZeroFlag(false)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	carryBit := env.regs[regA].get() & 0x01
	env.setCarryFlag(carryBit == 1)

	//fmt.Printf("RLCA\n")
	return 4
}

// rla rotates register A left by one, but uses the carry flag as a "bit 8" of
// sorts during this operation. This means that we're essentially rotating
// "(carry flag << 9) | register A".
func rla(env *environment) int {
	oldVal := uint8(env.regs[regA].get())
	// Get the current most significant bit, which will be put in the carry
	// flag
	var msb uint8
	if oldVal&0x80 == 0x80 {
		msb = 1
	} else {
		msb = 0
	}

	// Get the current carry bit, which will be put in the least significant
	// bit of register A
	var oldCarryVal uint8
	if env.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal << 1
	newVal |= oldCarryVal
	env.setCarryFlag(msb == 1)

	env.setZeroFlag(false)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	env.regs[regA].set(uint16(newVal))

	//fmt.Printf("RLA\n")
	return 4
}

// rrca bit rotates register A right by one, which is equivalent to a right bit
// shift where the least significant bit is carried over to the most
// significant bit. This bit is also stored in the carry flag.
func rrca(env *environment) int {
	rotated := bits.RotateLeft8(uint8(env.regs[regA].get()), -1)
	env.regs[regA].set(uint16(rotated))

	env.setZeroFlag(false)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	carryBit := env.regs[regA].get() & 0x80
	env.setCarryFlag(carryBit == 1)

	//fmt.Printf("RRCA\n")
	return 4
}

// rra rotates register A right by one, but uses the carry flag as a "bit -1"
// of sorts during this operation. This means that we're essentially rotating
// "carry flag | (register A << 1)".
func rra(env *environment) int {
	oldVal := uint8(env.regs[regA].get())
	// Get the current least significant bit, which will be put in the carry
	// flag
	var lsb uint8
	if oldVal&0x01 == 0x01 {
		lsb = 1
	} else {
		lsb = 0
	}

	// Get the current carry bit, which will be put in the most significant bit
	// of register A
	var oldCarryVal uint8
	if env.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal >> 1
	newVal |= (oldCarryVal << 7)
	env.setCarryFlag(lsb == 1)

	env.setZeroFlag(false)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	env.regs[regA].set(uint16(newVal))

	//fmt.Printf("RRA\n")
	return 4
}

// di sets the master interrupt flag to false, disabling all interrupt
// handling.
func di(env *environment) int {
	env.interruptsEnabled = false

	//fmt.Printf("DI\n")
	return 4
}

// di sets the master interrupt flag to true. Interrupts may still be disabled
// using the interrupt flags memory register, however.
func ei(env *environment) int {
	env.interruptsEnabled = true

	//fmt.Printf("EI\n")
	return 4
}

// halt stops running instructions until an interrupt is triggered.
func halt(env *environment) int {
	env.waitingForInterrupts = true

	//fmt.Printf("HALT\n")
	return 4
}

// cpl inverts the value of register A.
func cpl(env *environment) int {
	invertedA := ^uint8(env.regs[regA].get())
	env.regs[regA].set(uint16(invertedA))

	env.setHalfCarryFlag(true)
	env.setSubtractFlag(true)

	return 4
}

// ccf flips the carry flag.
func ccf(env *environment) int {
	env.setCarryFlag(!env.getCarryFlag())

	env.setHalfCarryFlag(false)
	env.setSubtractFlag(false)

	return 4
}

// scf sets the carry flag to true.
func scf(env *environment) int {
	env.setCarryFlag(true)
	env.setHalfCarryFlag(false)
	env.setSubtractFlag(false)

	return 4
}

// runOpcode runs the operation that maps to the given opcode.
func runOpcode(env *environment, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	//fmt.Printf("%#x: ", opcode)

	switch {
	// NOP
	case opcode == 0x00:
		return nop(env), nil
	// The CB prefix. Comes before an extended instruction set.
	case opcode == 0xCB:
		panic("CB prefix instructions are not supported")
	// STOP 0
	case opcode == 0x10:
		panic("STOP 0 is not supported")
	// DI
	case opcode == 0xF3:
		return di(env), nil
	// EI
	case opcode == 0xFB:
		return ei(env), nil
	// HALT
	case opcode == 0x76:
		return halt(env), nil
	// LD (BC),A
	case opcode == 0x02:
		return ldToMem(env, regBC, regA), nil
	// LD (DE),A
	case opcode == 0x12:
		return ldToMem(env, regDE, regA), nil
	// LD A,(BC)
	case opcode == 0x0A:
		return ldFromMem(env, regA, regBC), nil
	// LD A,(DE)
	case opcode == 0x1A:
		return ldFromMem(env, regA, regDE), nil
	// LDH (a8),A
	case opcode == 0xE0:
		return ldhToMem(env), nil
	// LDH A,(a8)
	case opcode == 0xF0:
		return ldhFromMem(env), nil
	// JR r8
	case opcode == 0x18:
		return jr(env), nil
	// JR NZ,r8
	case opcode == 0x20:
		return jrIfFlag(env, zeroFlag, false), nil
	// JR Z,r8
	case opcode == 0x28:
		return jrIfFlag(env, zeroFlag, true), nil
	// JR NC,r8
	case opcode == 0x30:
		return jrIfFlag(env, carryFlag, false), nil
	// JR C,r8
	case opcode == 0x30:
		return jrIfFlag(env, carryFlag, true), nil
	// JP a16
	case opcode == 0xC3:
		return jp(env), nil
	// RET
	case opcode == 0xC9:
		return ret(env), nil
	// RET NZ
	case opcode == 0xC0:
		return retIfFlag(env, zeroFlag, false), nil
	// RET Z
	case opcode == 0xC8:
		return retIfFlag(env, zeroFlag, true), nil
	// RET NC
	case opcode == 0xD0:
		return retIfFlag(env, carryFlag, false), nil
	// RET C
	case opcode == 0xD8:
		return retIfFlag(env, carryFlag, true), nil
	// RETI
	case opcode == 0xD9:
		return reti(env), nil
	// CALL address
	case opcode == 0xCD:
		return call(env), nil
	// RLCA
	case opcode == 0x07:
		return rlca(env), nil
	// RLA
	case opcode == 0x17:
		return rla(env), nil
	// RRCA
	case opcode == 0x0F:
		return rrca(env), nil
	// RRA
	case opcode == 0x1F:
		return rra(env), nil
	// LD A,(HL+)
	case opcode == 0x2A:
		return ldiFromMem(env), nil
	// LD A,(HL-)
	case opcode == 0x3A:
		return lddFromMem(env), nil
	// LD (HL+),A
	case opcode == 0x22:
		return ldiToMem(env), nil
	// LD (HL-),A
	case opcode == 0x32:
		return lddToMem(env), nil
	// ADD HL,nn
	case lowerNibble == 0x09 && upperNibble <= 0x03:
		return addToHL(env, indexTo16BitRegister[upperNibble]), nil
	// LD nn,SP
	case opcode == 0x08:
		return ldSPToMem(env), nil
	// LD r,d16
	case lowerNibble == 0x01 && upperNibble <= 0x03:
		return ld16BitImm(env, indexTo16BitRegister[upperNibble]), nil
	// LD A,(HL)
	case opcode == 0x7E:
		return ldFromMem(env, regA, regHL), nil
	// LD r,(HL)
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ldFromMem(env, indexToRegister[regIndex], regHL), nil
	// LD (HL),A
	case opcode == 0x77:
		return ldToMem(env, regHL, regA), nil
	// LD A,A
	case opcode == 0x7F:
		return ld(env, regA, regA), nil
	// LD r,A
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0F {
			regIndex++
		}
		return ld(env, indexToRegister[regIndex], regA), nil
	// LD B,r
	case upperNibble == 0x04 && lowerNibble <= 0x05:
		return ld(env, regB, indexToRegister[lowerNibble]), nil
	// LD C,r
	case upperNibble == 0x04 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regB, indexToRegister[regIndex]), nil
	// LD D,r
	case upperNibble == 0x05 && lowerNibble <= 0x05:
		return ld(env, regD, indexToRegister[lowerNibble]), nil
	// LD E,r
	case upperNibble == 0x05 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regE, indexToRegister[regIndex]), nil
	// LD H,r
	case upperNibble == 0x06 && lowerNibble <= 0x05:
		return ld(env, regH, indexToRegister[lowerNibble]), nil
	// LD L,r
	case upperNibble == 0x06 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regL, indexToRegister[regIndex]), nil
	// LD (HL),r
	case upperNibble == 0x07 && lowerNibble <= 0x05:
		return ldToMem(env, regHL, indexToRegister[lowerNibble]), nil
	// LD A,r
	case upperNibble == 0x07 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regA, indexToRegister[regIndex]), nil
	// LD (HL),d8
	case opcode == 0x36:
		return ld8BitImmToHLMem(env), nil
	// LD A,d8
	case opcode == 0x3E:
		return ld8BitImm(env, regA), nil
	// LD reg,d8
	case upperNibble < 0x02 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ld8BitImm(env, indexToRegister[regIndex]), nil
	// ADD A,A
	case opcode == 0x87:
		return add(env, regA), nil
	// ADD A,(HL)
	case opcode == 0x86:
		return addFromMemHL(env), nil
	// ADD A,r
	case opcode >= 0x80 && opcode <= 0x85:
		return add(env, indexToRegister[lowerNibble]), nil
	// SUB A
	case opcode == 0x97:
		return sub(env, regA), nil
	// SUB (HL)
	case opcode == 0x96:
		return subFromMemHL(env), nil
	// SUB r
	case opcode >= 0x90 && opcode <= 0x95:
		return sub(env, indexToRegister[lowerNibble]), nil
	// AND A
	case opcode == 0xA7:
		return and(env, regA), nil
	// AND (HL)
	case opcode == 0xA6:
		return andFromMemHL(env), nil
	// AND r
	case opcode >= 0xA0 && opcode <= 0xA5:
		return and(env, indexToRegister[lowerNibble]), nil
	// AND n
	case opcode == 0xE6:
		return and8BitImm(env), nil
	// INC (HL)
	case opcode == 0x34:
		return incMemHL(env), nil
	// INC A
	case opcode == 0x3C:
		return inc8Bit(env, regA), nil
	// INC r
	case upperNibble <= 0x02 && (lowerNibble == 0x04 || lowerNibble == 0x0C):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0C {
			regIndex++
		}

		return inc8Bit(env, indexToRegister[regIndex]), nil
	// INC ss
	case upperNibble <= 0x03 && lowerNibble == 0x03:
		return inc16Bit(env, indexTo16BitRegister[upperNibble]), nil
	// DEC (HL)
	case opcode == 0x35:
		return decMemHL(env), nil
	// DEC A
	case opcode == 0x3D:
		return dec8Bit(env, regA), nil
	// DEC r
	case upperNibble <= 0x02 && (lowerNibble == 0x05 || lowerNibble == 0x0D):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0D {
			regIndex += 1
		}

		return dec8Bit(env, indexToRegister[regIndex]), nil
	// DEC ss
	case upperNibble <= 0x03 && lowerNibble == 0x0B:
		return dec16Bit(env, indexTo16BitRegister[upperNibble]), nil
	// POP ss
	case upperNibble >= 0x0C && lowerNibble == 0x01:
		regIndex := upperNibble - 0x0C
		return pop(env, indexTo16BitRegister[regIndex]), nil
	// PUSH ss
	case upperNibble >= 0x0C && lowerNibble == 0x05:
		regIndex := upperNibble - 0x0C
		return push(env, indexTo16BitRegister[regIndex]), nil
	// RST address
	case upperNibble >= 0x0C && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		restartIndex := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			restartIndex++
		}

		return rst(env, indexToRestartAddress[restartIndex]), nil
	// CPL
	case opcode == 0x2F:
		return cpl(env), nil
	// CCF
	case opcode == 0x3F:
		return ccf(env), nil
	// SCF
	case opcode == 0x37:
		return scf(env), nil
	// ADD SP,r8
	case opcode == 0xE8:
		return addToSP(env), nil
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

// getConditionalStr creates a string representation of a conditional flag
// check.
func getConditionalStr(flagMask uint16, isSet bool) string {
	var conditional string
	switch flagMask {
	case zeroFlag:
		conditional = "Z"
	case carryFlag:
		conditional = "C"
	default:
		panic("unsupported JR conditional flag")
	}
	if !isSet {
		conditional = "N" + conditional
	}

	return conditional
}
