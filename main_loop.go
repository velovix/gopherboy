package main

import (
	"bufio"
	"fmt"
	"os"
)

// startMainLoop starts the main processing loop of the Gameboy.
func startMainLoop(env *environment, vc *videoController, timers *timers) error {
	stepping := false

	for {
		var err error

		var opTime int
		if env.waitingForInterrupts {
			// Spin our wheels running NOPs until an interrupt happens
			opTime = 4
		} else {
			// Fetch and run an operation
			opcode := env.incrementPC()

			if (breakOpcode != nil && opcode == *breakOpcode) || stepping {
				fmt.Printf("BREAK: %#02x\n", opcode)
				fmt.Printf("   A:   %#02x |  B:   %#02x\n",
					env.regs8[regA].get(),
					env.regs8[regB].get())
				fmt.Printf("   C:   %#02x |  D:   %#02x\n",
					env.regs8[regC].get(),
					env.regs8[regD].get())
				fmt.Printf("   E:   %#02x |  H:   %#02x\n",
					env.regs8[regE].get(),
					env.regs8[regH].get())
				fmt.Printf("   L:   %#02x |  F:   %#02x\n",
					env.regs8[regL].get(),
					env.regs8[regF].get())
				fmt.Printf("  AF: %#04x\n", env.regs16[regAF].get())
				fmt.Printf("  BC: %#04x\n", env.regs16[regBC].get())
				fmt.Printf("  DE: %#04x\n", env.regs16[regDE].get())
				fmt.Printf("  HL: %#04x\n", env.regs16[regHL].get())
				fmt.Printf("  SP: %#04x\n", env.regs16[regSP].get())
				fmt.Printf("  PC: %#04x\n", env.regs16[regPC].get()-1) // We've already incremented once to find the instruction

				reader := bufio.NewReader(os.Stdin)
				for {
					fmt.Print("Now what? ")
					command, _ := reader.ReadString('\n')
					if command == "c\n" {
						fmt.Println("Continuing")
						stepping = false
						break
					} else if command == "n\n" {
						fmt.Println("Stepping")
						stepping = true
						break
					} else {
						fmt.Printf("Unknown command '%v'\n", command)
						continue
					}
				}
			}

			opTime, err = runOpcode(env, opcode)
			if err != nil {
				return err
			}
		}

		// Process any delayed requests to toggle the master interrupt switch.
		// These are created by the EI and DI instructions.
		if env.enableInterruptsTimer > 0 {
			env.enableInterruptsTimer--
			if env.enableInterruptsTimer == 0 {
				env.interruptsEnabled = true
			}
		}
		if env.disableInterruptsTimer > 0 {
			env.disableInterruptsTimer--
			if env.disableInterruptsTimer == 0 {
				env.interruptsEnabled = false
			}
		}

		timers.tick(opTime)
		env.mmu.tick(opTime)
		vc.tick(opTime)

		// TODO(velovix): Should interrupt flags be unset here if the interrupt
		// is disabled?

		// Check if any interrupts need to be processed
		if env.interruptsEnabled && env.mmu.at(ifAddr) != 0 {
			var target uint16

			interruptEnable := env.mmu.at(ieAddr)
			interruptFlag := env.mmu.at(ifAddr)

			// Check each bit of the interrupt flag to see if an interrupt
			// happened, and each bit of the interrupt enable flag to check if
			// we should process it. Then, reset the interrupt flag.
			if interruptEnable&interruptFlag&0x01 == 0x01 {
				// VBlank interrupt
				target = vblankInterruptTarget
				interruptFlag &= ^uint8(0x01)
			} else if interruptEnable&interruptFlag&0x02 == 0x02 {
				// LCDC interrupt
				target = lcdcInterruptTarget
				interruptFlag &= ^uint8(0x02)
			} else if interruptEnable&interruptFlag&0x04 == 0x04 {
				// TIMA overflow interrupt
				target = timaOverflowInterruptTarget
				interruptFlag &= ^uint8(0x04)
			} else if interruptEnable&interruptFlag&0x08 == 0x08 {
				// Serial interrupt
				target = serialInterruptTarget
				interruptFlag &= ^uint8(0x08)
			} else if interruptEnable&interruptFlag&0x10 == 0x10 {
				// P1 thru P4 interrupt
				target = p1Thru4InterruptTarget
				interruptFlag &= ^uint8(0x10)
			}

			env.mmu.set(ifAddr, interruptFlag)

			if target != 0 {
				// Disable all other interrupts
				env.interruptsEnabled = false
				env.waitingForInterrupts = false
				// Push the current program counter to the stack for later use
				env.pushToStack16(env.regs16[regPC].get())
				// Jump to the target
				env.regs16[regPC].set(target)
			}
		}
	}
}
