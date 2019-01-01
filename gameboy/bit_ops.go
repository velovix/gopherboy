package gameboy

import "fmt"

// res creates an instruction that sets the specified bit of the given register
// to zero.
func res(bitNum uint8, reg registerType) instruction {
	return func(state *State) int {
		regVal := state.regs8[reg].get()

		state.regs8[reg].set(regVal & ^(0x1 << bitNum))

		if printInstructions {
			fmt.Printf("RES %v,%v\n", bitNum, reg)
		}
		return 8
	}
}

// resMemHL creates an instruction that sets the specified bit of the value at
// the address specified by register HL to zero.
func resMemHL(bitNum uint8) instruction {
	return func(state *State) int {
		hlVal := state.regs16[regHL].get()
		memVal := state.mmu.at(hlVal)

		memVal &= ^(0x1 << bitNum)
		state.mmu.set(hlVal, memVal)

		if printInstructions {
			fmt.Printf("RES %v,(%v)\n", bitNum, regHL)
		}
		return 16
	}
}

// bit creates an instruction that checks the given bit of the given register
// value.
func bit(bitNum uint8, reg registerType) instruction {
	return func(state *State) int {
		regVal := state.regs8[reg].get()

		bitSet := regVal&(0x1<<bitNum) == (0x1 << bitNum)

		state.setZeroFlag(!bitSet)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)

		if printInstructions {
			fmt.Printf("BIT %v,%v\n", bitNum, reg)
		}
		return 8
	}
}

// bitMemHL creates an instruction that checks the given bit of the value at
// the address specified by register HL.
func bitMemHL(bitNum uint8) instruction {
	return func(state *State) int {
		hlVal := state.regs16[regHL].get()
		memVal := state.mmu.at(hlVal)

		bitSet := memVal&(0x1<<bitNum) == (0x1 << bitNum)

		state.setZeroFlag(!bitSet)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(true)

		if printInstructions {
			fmt.Printf("BIT %v,(%v)\n", bitNum, regHL)
		}
		return 16
	}
}

// swap creates an instruction that swaps the upper and lower nibbles of the
// given register.
func swap(reg registerType) instruction {
	return func(state *State) int {
		regVal := state.regs8[reg].get()

		lower, upper := split(regVal)
		regVal = state.regs8[reg].set(combine(upper, lower))

		state.setZeroFlag(regVal == 0)
		state.setSubtractFlag(false)
		state.setHalfCarryFlag(false)
		state.setCarryFlag(false)

		if printInstructions {
			fmt.Printf("SWAP %v\n", reg)
		}

		return 8
	}
}

// swapMemHL swaps the upper and lower nibbles of the value in memory at the
// address specified by register HL.
func swapMemHL(state *State) int {
	hlVal := state.regs16[regHL].get()
	memVal := state.mmu.at(hlVal)

	lower, upper := split(memVal)
	memVal = combine(upper, lower)
	state.mmu.set(hlVal, memVal)

	state.setZeroFlag(memVal == 0)
	state.setSubtractFlag(false)
	state.setHalfCarryFlag(false)
	state.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("SWAP (%v)\n", regHL)
	}

	return 16
}

// set creates an instruction that sets the specified bit of the given register
// to one.
func set(bitNum uint8, reg registerType) instruction {
	return func(state *State) int {
		regVal := state.regs8[reg].get()

		state.regs8[reg].set(regVal | (0x1 << bitNum))

		if printInstructions {
			fmt.Printf("SET %v,%v\n", bitNum, reg)
		}
		return 8
	}
}

// setMemHL creates an instruction that sets the specified bit of the value at
// the address specified by register HL to one.
func setMemHL(bitNum uint8) instruction {
	return func(state *State) int {
		hlVal := state.regs16[regHL].get()
		memVal := state.mmu.at(hlVal)

		memVal |= 0x1 << bitNum
		state.mmu.set(hlVal, memVal)

		if printInstructions {
			fmt.Printf("SET %v,(%v)\n", bitNum, regHL)
		}
		return 16
	}
}
