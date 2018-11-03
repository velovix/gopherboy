package main

import "fmt"

// nop does nothing.
func nop(env *environment) int {
	if printInstructions {
		fmt.Printf("NOP\n")
	}
	return 4
}

// di sets the master interrupt flag to false, disabling all interrupt
// handling, but not until the instruction after DI has been executed.
func di(env *environment) int {
	env.disableInterruptsTimer = 2

	if printInstructions {
		fmt.Printf("DI\n")
	}
	return 4
}

// ei sets the master interrupt flag to true, but not until the instruction
// after EI has been executed. Interrupts may still be disabled using the
// interrupt flags memory register, however.
func ei(env *environment) int {
	env.enableInterruptsTimer = 2

	if printInstructions {
		fmt.Printf("EI\n")
	}
	return 4
}

// halt stops running instructions until an interrupt is triggered.
func halt(env *environment) int {
	env.waitingForInterrupts = true

	if printInstructions {
		fmt.Printf("HALT\n")
	}
	return 4
}

// cpl inverts the value of register A.
func cpl(env *environment) int {
	invertedA := ^env.regs8[regA].get()
	env.regs8[regA].set(invertedA)

	env.setHalfCarryFlag(true)
	env.setSubtractFlag(true)

	if printInstructions {
		fmt.Printf("CPL\n")
	}
	return 4
}

// ccf flips the carry flag.
func ccf(env *environment) int {
	env.setCarryFlag(!env.getCarryFlag())

	env.setHalfCarryFlag(false)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("CCF\n")
	}
	return 4
}

// scf sets the carry flag to true.
func scf(env *environment) int {
	env.setCarryFlag(true)
	env.setHalfCarryFlag(false)
	env.setSubtractFlag(false)

	if printInstructions {
		fmt.Printf("SCF\n")
	}
	return 4
}
