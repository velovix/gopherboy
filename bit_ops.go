package main

import "fmt"

// res sets the specified bit of the given register to zero.
func res(env *environment, bitNum uint8, reg registerType) int {
	regVal := env.regs[reg].get()

	env.regs[reg].set(regVal & ^(0x1 << bitNum))

	if printInstructions {
		fmt.Printf("RES %v,%v\n", bitNum, reg)
	}
	return 8
}

// swap swaps the upper and lower nibbles of the given register.
func swap(env *environment, reg registerType) int {
	regVal := env.regs[reg].get()

	lower, upper := split(uint8(regVal))
	regVal = env.regs[reg].set(uint16(combine(upper, lower)))

	env.setZeroFlag(regVal == 0)
	env.setSubtractFlag(false)
	env.setHalfCarryFlag(false)
	env.setCarryFlag(false)

	return 8
}
