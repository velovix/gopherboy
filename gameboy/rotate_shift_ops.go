package gameboy

import (
	"fmt"
	"math/bits"
)

// rlca bit rotates register A left by one, which is equivalent to a left bit
// shift where the most significant bit is carried over to the least
// significant bit. This bit is also stored in the carry flag.
func rlca(state *State) int {
	rotated := bits.RotateLeft8(state.regs8[regA].get(), 1)
	state.regs8[regA].set(rotated)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	carryBit := state.regs8[regA].get() & 0x01
	state.setCarryFlag(carryBit == 1)

	if printInstructions {
		fmt.Printf("RLCA\n")
	}
	return 4
}

// rla rotates register A left by one, but uses the carry flag as a "bit 8" of
// sorts during this operation. This means that we're essentially rotating
// "(carry flag << 1) | register A".
func rla(state *State) int {
	oldVal := state.regs8[regA].get()
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
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal << 1
	newVal |= oldCarryVal
	state.setCarryFlag(msb == 1)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.regs8[regA].set(newVal)

	if printInstructions {
		fmt.Printf("RLA\n")
	}
	return 4
}

// rrca bit rotates register A right by one, which is equivalent to a right bit
// shift where the least significant bit is carried over to the most
// significant bit. This bit is also stored in the carry flag.
func rrca(state *State) int {
	carryBit := state.regs8[regA].get() & 0x01
	state.setCarryFlag(carryBit == 1)

	rotated := bits.RotateLeft8(state.regs8[regA].get(), -1)
	state.regs8[regA].set(rotated)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("RRCA\n")
	}
	return 4
}

// rra rotates register A right by one, but uses the carry flag as a "bit -1"
// of sorts during this operation. This means that we're essentially rotating
// "carry flag | (register A << 1)".
func rra(state *State) int {
	oldVal := state.regs8[regA].get()
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
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal >> 1
	newVal |= (oldCarryVal << 7)
	state.setCarryFlag(lsb == 1)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.regs8[regA].set(newVal)

	if printInstructions {
		fmt.Printf("RRA\n")
	}
	return 4
}

// srl shifts the contents of the given register to the right. Bit 0 is shifted
// to the carry register. Bit 7 is set to 0.
func srl(state *State, reg registerType) int {
	regVal := state.regs8[reg].get()

	// Put the least significant bit in the carry register
	lsb := regVal & 0x01
	state.setCarryFlag(lsb == 1)

	regVal = state.regs8[reg].set(regVal >> 1)

	state.setZeroFlag(regVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SRL %v\n", reg)
	}
	return 8
}

// srlMemHL shifts the value at the address in memory specified by register
// HL to the right. Bit 0 is shifted to the carry register. Bit 7 is set to 0.
func srlMemHL(state *State) int {
	hlVal := state.regs16[regHL].get()
	memVal := state.mmu.at(hlVal)

	// Put the least significant bit in the carry register
	lsb := memVal & 0x01
	state.setCarryFlag(lsb == 1)

	state.mmu.set(hlVal, memVal>>1)

	state.setZeroFlag(memVal>>1 == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SRL (%v)\n", regHL)
	}

	return 16
}

// rr rotates the contents of the given register right by one, but uses the
// carry flag as a "bit -1" of sorts during this operation. This means we're
// essentially rotating "(register << 1) | carry flag".
func rr(state *State, reg registerType) int {
	oldVal := state.regs8[reg].get()
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
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal >> 1
	newVal |= (oldCarryVal << 7)
	state.setCarryFlag(lsb == 1)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.regs8[reg].set(newVal)

	if printInstructions {
		fmt.Printf("RR %v\n", reg)
	}
	return 8
}

// rrMemHL rotates the value stored in memory at the address specified by
// register HL by 1. The carry flag is used as a "bit -1" of sorts during this
// operation. This means we're essentially rotating
// "(mem[regHL] << 1) | carryFlag".
func rrMemHL(state *State) int {
	hlVal := state.regs16[regHL].get()
	oldVal := state.mmu.at(hlVal)

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
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal >> 1
	newVal |= (oldCarryVal << 7)
	state.setCarryFlag(lsb == 1)

	state.setZeroFlag(newVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.mmu.set(hlVal, newVal)

	if printInstructions {
		fmt.Printf("RR (%v)\n", regHL)
	}
	return 16
}

// rlc bit rotates the given register left by one, which is equivalent to a
// left bit shift where the most significant bit is carried over to the least
// significant bit. This bit is also stored in the carry flag.
func rlc(state *State, reg registerType) int {
	rotated := bits.RotateLeft8(state.regs8[reg].get(), 1)
	state.regs8[reg].set(rotated)

	state.setZeroFlag(rotated == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	carryBit := state.regs8[reg].get() & 0x01
	state.setCarryFlag(carryBit == 1)

	if printInstructions {
		fmt.Printf("RLC %v\n", reg)
	}
	return 8
}

// rlcMemHL bit rotates the value found in memory at the address specified by
// HL left by one, which is equivalent to a left bit shift where the most
// significant bit is carried over to the least significant bit. This bit is
// also stored in the carry flag.
func rlcMemHL(state *State) int {
	memVal := state.mmu.at(state.regs16[regHL].get())

	memVal = bits.RotateLeft8(memVal, 1)
	state.mmu.set(state.regs16[regHL].get(), memVal)

	state.setZeroFlag(memVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	carryBit := memVal & 0x01
	state.setCarryFlag(carryBit == 1)

	if printInstructions {
		fmt.Printf("RLC (%v)\n", regHL)
	}
	return 16
}

// rrc bit rotates the given register right by one, which is equivalent to a
// right bit shift where the least significant bit is carried over to the most
// significant bit. This bit is also stored in the carry flag.
func rrc(state *State, reg registerType) int {
	carryBit := state.regs8[reg].get() & 0x01
	state.setCarryFlag(carryBit == 1)

	rotated := bits.RotateLeft8(state.regs8[reg].get(), -1)
	state.regs8[reg].set(rotated)

	state.setZeroFlag(rotated == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("RRC %v\n", reg)
	}
	return 8
}

// rrcMemHL bit rotates the value found in memory at the address specified by
// HL right by one, which is equivalent to a right bit shift where the least
// significant bit is carried over to the most significant bit. This bit is
// also stored in the carry flag.
func rrcMemHL(state *State) int {
	memVal := state.mmu.at(state.regs16[regHL].get())

	carryBit := memVal & 0x01

	memVal = bits.RotateLeft8(memVal, -1)
	state.mmu.set(state.regs16[regHL].get(), memVal)

	state.setZeroFlag(memVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.setCarryFlag(carryBit == 1)

	if printInstructions {
		fmt.Printf("RRC (%v)\n", regHL)
	}
	return 16
}

// rl rotates the given register value left by one, but uses the carry flag as
// a "bit 8" of sorts during this operation. This means that we're essentially
// rotating "(carry flag << 1) | register A".
func rl(state *State, reg registerType) int {
	oldVal := state.regs8[reg].get()
	// Get the current most significant bit, which will be put in the carry
	// flag
	var msb uint8
	if oldVal&0x80 == 0x80 {
		msb = 1
	} else {
		msb = 0
	}

	// Get the current carry bit, which will be put in the least significant
	// bit of the register
	var oldCarryVal uint8
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal << 1
	newVal |= oldCarryVal
	state.setCarryFlag(msb == 1)

	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.regs8[reg].set(newVal)

	state.setZeroFlag(newVal == 0)

	if printInstructions {
		fmt.Printf("RL %v\n", reg)
	}
	return 8
}

// rlMemHL rotates the value in memory at the address specified by register HL
// left by one, but uses the carry flag as a "bit 8" of sorts during this
// operation. This means that we're essentially rotating
// "(carry flag << 1) | mem(regHL)".
func rlMemHL(state *State) int {
	oldVal := state.mmu.at(state.regs16[regHL].get())
	// Get the current most significant bit, which will be put in the carry
	// flag
	var msb uint8
	if oldVal&0x80 == 0x80 {
		msb = 1
	} else {
		msb = 0
	}

	// Get the current carry bit, which will be put in the least significant
	// bit of the register
	var oldCarryVal uint8
	if state.getCarryFlag() {
		oldCarryVal = 1
	} else {
		oldCarryVal = 0
	}

	newVal := oldVal << 1
	newVal |= oldCarryVal
	state.setCarryFlag(msb == 1)

	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	state.mmu.set(state.regs16[regHL].get(), newVal)

	state.setZeroFlag(newVal == 0)

	if printInstructions {
		fmt.Printf("RL (%v)\n", regHL)
	}
	return 16
}

// sla shifts the contents of the given register to the left. Bit 7 is shifted
// to the carry register. Bit 0 is set to 0.
func sla(state *State, reg registerType) int {
	regVal := state.regs8[reg].get()

	// Put the most significant bit in the carry register
	msb := regVal&0x80 == 0x80
	state.setCarryFlag(msb)

	regVal = state.regs8[reg].set(regVal << 1)

	state.setZeroFlag(regVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SLA %v\n", reg)
	}
	return 8
}

// slaMemHL shifts the value at the address in memory specified by register
// HL to the left. Bit 7 is shifted to the carry register. Bit 0 is set to 0.
func slaMemHL(state *State) int {
	hlVal := state.regs16[regHL].get()
	memVal := state.mmu.at(hlVal)

	// Put the most significant bit in the carry register
	state.setCarryFlag(memVal&0x80 == 0x80)

	memVal <<= 1
	state.mmu.set(hlVal, memVal)

	state.setZeroFlag(memVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SLA (%v)\n", regHL)
	}
	return 16
}

// sra shifts the contents of the given register to the right. Bit 0 is shifted
// to the carry register. Bit 7 is left unchanged.
func sra(state *State, reg registerType) int {
	regVal := state.regs8[reg].get()

	// Put the least significant bit in the carry register
	lsb := regVal & 0x01
	state.setCarryFlag(lsb == 1)

	msb := regVal & 0x80

	regVal >>= 1

	// Put the previous most significant bit back in bit 7
	regVal |= msb
	regVal = state.regs8[reg].set(regVal)

	state.setZeroFlag(regVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SRA %v\n", reg)
	}
	return 8
}

// sraMemHL shifts the value at the address in memory specified by register HL
// to the right. Bit 0 is shifted to the carry register. Bit 7 is unchanged.
func sraMemHL(state *State) int {
	hlVal := state.regs16[regHL].get()
	memVal := state.mmu.at(hlVal)

	// Put the least significant bit in the carry register
	lsb := memVal & 0x01
	state.setCarryFlag(lsb == 1)

	memVal = memVal >> 1

	// Put the previous most significant bit back in bit 7
	memVal |= (memVal & 0x40) << 1

	state.mmu.set(hlVal, memVal)

	state.setZeroFlag(memVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	if printInstructions {
		fmt.Printf("SRA (%v)\n", regHL)
	}
	return 16
}