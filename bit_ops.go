package main

import "fmt"

// res sets the specified bit of the given register to zero.
func res(env *environment, bitNum uint8, reg registerType) int {
	regVal := env.regs8[reg].get()

	env.regs8[reg].set(regVal & ^(0x1 << bitNum))

	if printInstructions {
		fmt.Printf("RES %v,%v\n", bitNum, reg)
	}
	return 8
}

// resMemHL sets the specified bit of the value at the address specified by
// register HL to zero.
func resMemHL(env *environment, bitNum uint8) int {
	hlVal := env.regs16[regHL].get()
	memVal := env.mmu.at(hlVal)

	memVal &= ^(0x1 << bitNum)
	env.mmu.set(hlVal, memVal)

	if printInstructions {
		fmt.Printf("RES %v,(%v)\n", bitNum, regHL)
	}
	return 16
}

// bit checks the given bit of the given register value.
func bit(env *environment, bitNum uint8, reg registerType) int {
	regVal := env.regs8[reg].get()

	bitSet := regVal&(0x1<<bitNum) == (0x1 << bitNum)

	env.setZeroFlag(!bitSet)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)

	if printInstructions {
		fmt.Printf("BIT %v,%v\n", bitNum, reg)
	}
	return 8
}

// bitMemHL checks the given bit of the value at the address specified by
// register HL.
func bitMemHL(env *environment, bitNum uint8) int {
	hlVal := env.regs16[regHL].get()
	memVal := env.mmu.at(hlVal)

	bitSet := memVal&(0x1<<bitNum) == (0x1 << bitNum)

	env.setZeroFlag(!bitSet)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(true)

	if printInstructions {
		fmt.Printf("BIT %v,(%v)\n", bitNum, regHL)
	}
	return 16
}

// swap swaps the upper and lower nibbles of the given register.
func swap(env *environment, reg registerType) int {
	regVal := env.regs8[reg].get()

	lower, upper := split(regVal)
	regVal = env.regs8[reg].set(combine(upper, lower))

	env.setZeroFlag(regVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("SWAP %v\n", reg)
	}

	return 8
}

// swapMemHL swaps the upper and lower nibbles of the value in memory at the
// address specified by register HL.
func swapMemHL(env *environment) int {
	hlVal := env.regs16[regHL].get()
	memVal := env.mmu.at(hlVal)

	lower, upper := split(memVal)
	memVal = combine(upper, lower)
	env.mmu.set(hlVal, memVal)

	env.setZeroFlag(memVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	if printInstructions {
		fmt.Printf("SWAP (%v)\n", regHL)
	}

	return 16
}

// set sets the specified bit of the given register to one.
func set(env *environment, bitNum uint8, reg registerType) int {
	regVal := env.regs8[reg].get()

	env.regs8[reg].set(regVal | (0x1 << bitNum))

	if printInstructions {
		fmt.Printf("SET %v,%v\n", bitNum, reg)
	}
	return 8
}

// setMemHL sets the specified bit of the value at the address specified by
// register HL to one.
func setMemHL(env *environment, bitNum uint8) int {
	hlVal := env.regs16[regHL].get()
	memVal := env.mmu.at(hlVal)

	memVal |= 0x1 << bitNum
	env.mmu.set(hlVal, memVal)

	if printInstructions {
		fmt.Printf("RES %v,(%v)\n", bitNum, regHL)
	}
	return 16
}
