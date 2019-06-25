package gameboy

import (
	"math/bits"
)

// rlca bit rotates register A left by one, which is equivalent to a left bit
// shift where the most significant bit is carried over to the least
// significant bit. This bit is also stored in the carry flag.
func rlca(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	rotated := bits.RotateLeft8(state.regA.get(), 1)
	state.regA.set(rotated)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	carryBit := state.regA.get() & 0x01
	state.setCarryFlag(carryBit == 1)

	return nil
}

// rla rotates register A left by one, but uses the carry flag as a "bit 8" of
// sorts during this operation. This means that we're essentially rotating
// "(carry flag << 1) | register A".
func rla(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	oldVal := state.regA.get()
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

	state.regA.set(newVal)

	return nil
}

// rrca bit rotates register A right by one, which is equivalent to a right bit
// shift where the least significant bit is carried over to the most
// significant bit. This bit is also stored in the carry flag.
func rrca(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	carryBit := state.regA.get() & 0x01
	state.setCarryFlag(carryBit == 1)

	rotated := bits.RotateLeft8(state.regA.get(), -1)
	state.regA.set(rotated)

	state.setZeroFlag(false)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)

	return nil
}

// rra rotates register A right by one, but uses the carry flag as a "bit -1"
// of sorts during this operation. This means that we're essentially rotating
// "carry flag | (register A << 1)".
func rra(state *State) instruction {
	// M-Cycle 0: Fetch instruction and do operation

	oldVal := state.regA.get()
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

	state.regA.set(newVal)

	return nil
}

// makeSRL creates an instruction that shifts the contents of the given
// register to the right. Bit 0 is shifted to the carry register. Bit 7 is set
// to 0.
func makeSRL(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		regVal := reg.get()

		// Put the least significant bit in the carry register
		lsb := regVal & 0x01
		state.setCarryFlag(lsb == 1)

		regVal = reg.set(regVal >> 1)

		state.setZeroFlag(regVal == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)

		return nil
	}
}

// srlMemHL shifts the value at the address in memory specified by register
// HL to the right. Bit 0 is shifted to the carry register. Bit 7 is set to 0.
func srlMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		hlVal := state.regHL.get()
		memVal := state.mmu.at(hlVal)

		return func(state *State) instruction {
			// M-Cycle 3: Do operation and write to memory at HL

			// Put the least significant bit in the carry register
			lsb := memVal & 0x01
			state.setCarryFlag(lsb == 1)

			state.mmu.set(hlVal, memVal>>1)

			state.setZeroFlag(memVal>>1 == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(false)

			return nil
		}
	}
}

// makeRR creates an instruction that rotates the contents of the given
// register right by one, but uses the carry flag as a "bit -1" of sorts during
// this operation. This means we're essentially rotating "(register << 1) |
// carry flag".
func makeRR(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		oldVal := reg.get()
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

		reg.set(newVal)

		return nil
	}
}

// rrMemHL rotates the value stored in memory at the address specified by
// register HL by 1. The carry flag is used as a "bit -1" of sorts during this
// operation. This means we're essentially rotating
// "(mem[regHL] << 1) | carryFlag".
func rrMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		hlVal := state.regHL.get()
		oldVal := state.mmu.at(hlVal)

		return func(state *State) instruction {
			// M-Cycle 3: Do operation and write to memory at HL

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

			return nil
		}
	}
}

// makeRLC creates an instruction that bit rotates the given register left by
// one, which is equivalent to a left bit shift where the most significant bit
// is carried over to the least significant bit. This bit is also stored in the
// carry flag.
func makeRLC(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		rotated := bits.RotateLeft8(reg.get(), 1)
		reg.set(rotated)

		state.setZeroFlag(rotated == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)

		carryBit := reg.get() & 0x01
		state.setCarryFlag(carryBit == 1)

		return nil
	}
}

// rlcMemHL bit rotates the value found in memory at the address specified by
// HL left by one, which is equivalent to a left bit shift where the most
// significant bit is carried over to the least significant bit. This bit is
// also stored in the carry flag.
func rlcMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from HL location in memory

		memVal := state.mmu.at(state.regHL.get())

		return func(state *State) instruction {
			// M-Cycle 2: Write to HL location in memory

			memVal = bits.RotateLeft8(memVal, 1)
			state.mmu.set(state.regHL.get(), memVal)

			state.setZeroFlag(memVal == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(false)

			carryBit := memVal & 0x01
			state.setCarryFlag(carryBit == 1)

			return nil
		}
	}
}

// makeRRC creates an instruction that bit rotates the given register right by
// one, which is equivalent to a right bit shift where the least significant
// bit is carried over to the most significant bit. This bit is also stored in
// the carry flag.
func makeRRC(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		carryBit := reg.get() & 0x01
		state.setCarryFlag(carryBit == 1)

		rotated := bits.RotateLeft8(reg.get(), -1)
		reg.set(rotated)

		state.setZeroFlag(rotated == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)

		return nil
	}
}

// rrcMemHL bit rotates the value found in memory at the address specified by
// HL right by one, which is equivalent to a right bit shift where the least
// significant bit is carried over to the most significant bit. This bit is
// also stored in the carry flag.
func rrcMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		memVal := state.mmu.at(state.regHL.get())

		return func(state *State) instruction {
			// M-Cycle 3: Do operation and write to memory at HL

			carryBit := memVal & 0x01

			memVal = bits.RotateLeft8(memVal, -1)
			state.mmu.set(state.regHL.get(), memVal)

			state.setZeroFlag(memVal == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(false)

			state.setCarryFlag(carryBit == 1)

			return nil
		}

	}
}

// makeRL creates an instruction that rotates the given register value left by
// one, but uses the carry flag as a "bit 8" of sorts during this operation.
// This means that we're essentially rotating "(carry flag << 1) | register A".
func makeRL(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		oldVal := reg.get()
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

		reg.set(newVal)

		state.setZeroFlag(newVal == 0)

		return nil
	}
}

// rlMemHL rotates the value in memory at the address specified by register HL
// left by one, but uses the carry flag as a "bit 8" of sorts during this
// operation. This means that we're essentially rotating
// "(carry flag << 1) | mem(regHL)".
func rlMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		oldVal := state.mmu.at(state.regHL.get())

		return func(state *State) instruction {
			// M-Cycle 3: Write to memory at HL

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

			state.mmu.set(state.regHL.get(), newVal)

			state.setZeroFlag(newVal == 0)

			return nil
		}
	}
}

// makeSLA creates an instruction that shifts the contents of the given
// register to the left. Bit 7 is shifted to the carry register. Bit 0 is set
// to 0.
func makeSLA(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		regVal := reg.get()

		// Put the most significant bit in the carry register
		msb := regVal&0x80 == 0x80
		state.setCarryFlag(msb)

		regVal = reg.set(regVal << 1)

		state.setZeroFlag(regVal == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)

		return nil
	}
}

// slaMemHL shifts the value at the address in memory specified by register
// HL to the left. Bit 7 is shifted to the carry register. Bit 0 is set to 0.
func slaMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		hlVal := state.regHL.get()
		memVal := state.mmu.at(hlVal)

		return func(state *State) instruction {
			// M-Cycle 3: Write to memory at HL

			// Put the most significant bit in the carry register
			state.setCarryFlag(memVal&0x80 == 0x80)

			memVal <<= 1
			state.mmu.set(hlVal, memVal)

			state.setZeroFlag(memVal == 0)
			state.setSubtractFlag(false)
			state.setHalfCarryFlag(false)

			return nil
		}
	}
}

// makeSRA creates an instruction that shifts the contents of the given
// register to the right. Bit 0 is shifted to the carry register. Bit 7 is left
// unchanged.
func makeSRA(reg register8) instruction {
	return func(state *State) instruction {
		// M-Cycle 1: Do operation

		regVal := reg.get()

		// Put the least significant bit in the carry register
		lsb := regVal & 0x01
		state.setCarryFlag(lsb == 1)

		msb := regVal & 0x80

		regVal >>= 1

		// Put the previous most significant bit back in bit 7
		regVal |= msb
		regVal = reg.set(regVal)

		state.setZeroFlag(regVal == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)

		return nil
	}
}

// sraMemHL shifts the value at the address in memory specified by register HL
// to the right. Bit 0 is shifted to the carry register. Bit 7 is unchanged.
func sraMemHL(state *State) instruction {
	// M-Cycle 1: Fetch CB instruction

	return func(state *State) instruction {
		// M-Cycle 2: Read from memory at HL

		hlVal := state.regHL.get()
		memVal := state.mmu.at(hlVal)

		return func(state *State) instruction {
			// M-Cycle 3: Do operation and write to memory at HL

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

			return nil
		}
	}
}
