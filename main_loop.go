package main

// startMainLoop starts the main processing loop of the Gameboy.
func startMainLoop(env *environment, vc *videoController, timers *timers) error {
	for {
		var err error

		var opTime int
		if env.waitingForInterrupts {
			// Spin our wheels running NOPs until an interrupt happens
			opTime = 4
		} else {
			// Fetch and run an operation
			opcode := env.incrementPC()
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
				env.pushToStack16(env.regs[regPC].get())
				// Jump to the target
				env.regs[regPC].set(target)
			}
		}
	}
}