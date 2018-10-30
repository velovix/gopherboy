package main

import (
	"fmt"
	"math/bits"
)

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

	if printInstructions {
		fmt.Printf("RLCA\n")
	}
	return 4
}

// rla rotates register A left by one, but uses the carry flag as a "bit 8" of
// sorts during this operation. This means that we're essentially rotating
// "(carry flag << 1) | register A".
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

	if printInstructions {
		fmt.Printf("RLA\n")
	}
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

	if printInstructions {
		fmt.Printf("RRCA\n")
	}
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

	if printInstructions {
		fmt.Printf("RRA\n")
	}
	return 4
}

// srl shifts the contents of the given register to the right. Bit 0 is shifted
// to the carry register. Bit 7 is set to 0.
func srl(env *environment, reg registerType) int {
	regVal := env.regs[reg].get()

	// Put the least significant bit in the carry register
	lsb := regVal & 0x01
	env.setCarryFlag(lsb == 1)

	regVal = env.regs[reg].set(regVal >> 1)

	env.setZeroFlag(regVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	return 8
}

// srlMemHL shifts the value at the address in memory specified by register
// HL to the right. Bit 0 is shifted to the carry register. Bit 7 is set to 0.
func srlMemHL(env *environment) int {
	hlVal := env.regs[regHL].get()
	memVal := env.mmu.at(hlVal)

	// Put the least significant bit in the carry register
	lsb := memVal & 0x01
	env.setCarryFlag(lsb == 1)

	env.mmu.set(hlVal, memVal>>1)

	env.setZeroFlag(memVal>>1 == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	return 16
}

// rr rotates the contents of the given register right by one, but uses the
// carry flag as a "bit -1" of sorts during this operation. This means we're
// essentially rotating "(register << 1) | carry flag".
func rr(env *environment, reg registerType) int {
	oldVal := uint8(env.regs[reg].get())
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

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	env.regs[reg].set(uint16(newVal))

	if printInstructions {
		fmt.Printf("RR %v\n", reg)
	}
	return 4
}

// rrMemHL rotates the value stored in memory at the address specified by
// register HL by 1. The carry flag is used as a "bit -1" of sorts during this
// operation. This means we're essentially rotating
// "(mem[regHL] << 1) | carryFlag".
func rrMemHL(env *environment) int {
	hlVal := env.regs[regHL].get()
	oldVal := env.mmu.at(hlVal)

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

	env.setZeroFlag(newVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)

	env.mmu.set(hlVal, newVal)

	if printInstructions {
		fmt.Printf("RR (HL)\n")
	}
	return 4
}
