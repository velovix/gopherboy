package main

const mClockRate = 1048576

func startMainLoop(env *environment) error {
	// A clock that increments for every clock cycle
	tClock := 0
	// A clock that increments for every 4 clock cycles
	mClock := 0
	// A clock that increments for every 64 clock cycles
	divider := &env.mem[dividerAddr]
	// TIMA is a configurable timer. Also known as the "counter".
	tima := &env.mem[timaAddr]
	// The value that the TIMA should start back at after overflowing. Also
	// known as the "modulo".
	tma := &env.mem[tmaAddr]
	// A byte that the user can use to control the TIMA. Also known as
	// "control".
	tac := &env.mem[tacAddr]

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

		timaRate, timaRunning := parseTAC(*tac)

		// Increment the clock based on how long the operation lasted
		for i := 0; i < opTime; i++ {
			tClock++
			if tClock%4 == 0 {
				mClock++
			}
			if mClock%64 == 0 {
				*divider++
			}
			// Finds how many m-clock increments should happen before a
			// TIMA increment should happen.
			clocksPerTimer := mClockRate / timaRate
			if timaRunning && mClock%clocksPerTimer == 0 {
				*tima++
				if *tima == 0 {
					// Flag a TIMA overflow interrupt
					env.mem[ieAddr] |= 0x04
					// Start back up at the specified modulo value
					*tima = *tma
				}
			}
		}

		// Check if any interrupts need to be processed
		if env.interruptsEnabled && env.mem[ifAddr] != 0 {
			var target uint16

			interruptEnable := env.mem[ieAddr]
			interruptFlag := env.mem[ifAddr]

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

			if target != 0 {
				env.waitingForInterrupts = false
				// Push the current program counter to the stack for later use
				env.push(regPC)
				// Jump to the target
				env.regs[regPC].set(target)
			}
		}
	}
}

// parseTAC takes in a control byte and returns the configuration it supplies.
// The rate refers to the rate at which the TIMA should run. The running value
// refers to whether or not the TIMA should run in the first place.
func parseTAC(tac uint8) (rate int, running bool) {
	speedBits := tac & 0x3
	switch speedBits {
	case 0x0:
		rate = 4096
	case 0x1:
		rate = 262144
	case 0x2:
		rate = 65536
	case 0x3:
		rate = 16384
	}

	runningBit := tac & 0x4
	return rate, runningBit == 0x4
}
