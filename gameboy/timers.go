package gameboy

import (
	"time"
)

const (
	// The Game Boy processor clock speed
	cpuClockRate = 4194304

	// The time that it takes in hardware to perform one cycle.
	timePerClock = time.Nanosecond * 238
)

// timers keeps track of all timers in the Gameboy, including the TIMA.
type timers struct {
	// A clock that increments every clock cycle
	cpuClock uint16

	// The last value of a CPU clock bit. Used to detect falling edges in the
	// system clock, which trigger a TIMA increment.
	timaDelay uint8

	state *State
}

func newTimers(state *State) *timers {
	t := &timers{state: state}

	t.state.mmu.subscribeTo(dividerAddr, t.onDividerWrite)
	t.state.mmu.subscribeTo(tacAddr, t.onTACWrite)

	// The CPU runs 2 NOPs before the boot ROM starts. Fake these NOPs by
	// incrementing the timer.
	// TODO(velovix): Is this sufficient or do these NOPs effect any other
	// subsystem?
	t.tick(8)

	return t
}

// tick increments the timers given the amount of cycles that have passed since
// the last call to tick. Flags interrupts as needed.
func (t *timers) tick(amount int) {
	// Parse the TAC bits for TIMA configuration information
	tac := t.state.mmu.at(tacAddr)
	timaRunning := tac&0x4 == 0x4
	timaRateBits := tac & 0x3

	// Load the current TIMA value
	tima := t.state.mmu.at(timaAddr)

	for i := 0; i < amount; i++ {
		t.cpuClock++

		// Pull the bit of interest from the CPU clock
		var timaBit uint8
		switch timaRateBits {
		case 0x0:
			// TIMA configured at 4096 Hz
			timaBit = uint8((t.cpuClock >> 9) & 0x1)
		case 0x3:
			// TIMA configured at 16384 Hz
			timaBit = uint8((t.cpuClock >> 7) & 0x1)
		case 0x2:
			// TIMA configured at 65536 Hz
			timaBit = uint8((t.cpuClock >> 5) & 0x1)
		case 0x1:
			// TIMA configured at 262144 Hz
			timaBit = uint8((t.cpuClock >> 3) & 0x1)
		}

		// Detect a falling edge on this bit and increment the TIMA if one is
		// detected
		var timaBitAndEnabled uint8
		if timaBit == 1 && timaRunning {
			timaBitAndEnabled = 1
		}
		timaShouldIncrement := timaBitAndEnabled != 1 && t.timaDelay == 1
		if timaShouldIncrement {
			tima++
			if tima == 0 {
				// Start back up at the specified modulo value
				tima = t.state.mmu.at(tmaAddr)

				timaInterruptEnabled := t.state.mmu.at(ieAddr)&0x04 == 0x04
				if t.state.interruptsEnabled && timaInterruptEnabled {
					// Flag a TIMA overflow interrupt
					t.state.mmu.setNoNotify(ifAddr, t.state.mmu.at(ifAddr)|0x04)
				}
			}
		}
		t.timaDelay = timaBitAndEnabled
	}

	// Update the timers in memory
	// The divider register is simply the 8 most significant bits of the CPU
	// clock
	t.state.mmu.setNoNotify(dividerAddr, uint8(t.cpuClock>>8))
	t.state.mmu.setNoNotify(timaAddr, tima)
}

// onDividerWrite is called when the divider register is written to. This
// triggers the divider timer to reset to zero.
func (t *timers) onDividerWrite(addr uint16, writeVal uint8) uint8 {
	t.cpuClock = 0
	return 0
}

// onTACWrite is called when the TAC register is written to. This controls
// various aspects of the TIMA timer.
func (t *timers) onTACWrite(addr uint16, writeVal uint8) uint8 {
	// This register is only 3 bits in size, get those bits
	writeVal = writeVal & 0x07

	// All unused bits are high
	return 0xF8 | writeVal
}
