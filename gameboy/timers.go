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
	cpuClock uint32

	// The last value of a CPU clock bit. Used to detect falling edges in the
	// system clock, which trigger a TIMA increment.
	timaDelay uint8

	// A countdown from when the TIMA originally overflowed to when the
	// interrupt should happen and when the TIMA should be set to the TMA
	// value. This is used to emulate the delayed TIMA interrupt behavior.
	timaInterruptCountdown int
	// When this value is greater than 0, the process of transferring the TMA
	// value into the TIMA is happening. During this time, the TIMA is
	// protected from writes and any updates to the TMA are immediately
	// reflected in the TIMA.
	tmaToTIMATransferCountdown int

	state *State
}

func newTimers(state *State) *timers {
	t := &timers{state: state}

	t.state.mmu.subscribeTo(dividerAddr, t.onDividerWrite)
	t.state.mmu.subscribeTo(tacAddr, t.onTACWrite)
	t.state.mmu.subscribeTo(timaAddr, t.onTIMAWrite)
	t.state.mmu.subscribeTo(tmaAddr, t.onTMAWrite)

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
	tac := t.state.mmu.atIORAM(tacAddr)
	timaRunning := tac&0x4 == 0x4
	timaRateBits := tac & 0x3

	// Load the current TIMA value
	tima := t.state.mmu.atIORAM(timaAddr)

	for i := 0; i < amount; i++ {
		t.cpuClock++
		if t.cpuClock == cpuClockRate {
			t.cpuClock = 0
		}
		// TODO(velovix): Should we do a manual wraparound here?

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

		// Count down the TIMA protection
		if t.tmaToTIMATransferCountdown > 0 {
			t.tmaToTIMATransferCountdown--
		}

		// Check the countdown to see if it's time to process a delayed TIMA
		// interrupt
		if t.timaInterruptCountdown > 0 {
			t.timaInterruptCountdown--
			if t.timaInterruptCountdown == 0 {
				// Start back up at the specified modulo value
				tima = t.state.mmu.atIORAM(tmaAddr)

				timaInterruptEnabled := t.state.mmu.atHRAM(ieAddr)&0x04 == 0x04
				if t.state.interruptsEnabled && timaInterruptEnabled {
					// Flag a TIMA overflow interrupt
					t.state.mmu.setIORAM(ifAddr, t.state.mmu.atIORAM(ifAddr)|0x04)
				}

				// The TMA to TIMA transfer process has been initiated
				t.tmaToTIMATransferCountdown = 4
			}
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
				// There is a 4-cycle delay between the TIMA overflow and the
				// interrupt and reset to TMA
				t.timaInterruptCountdown = 4
			}
		}
		t.timaDelay = timaBitAndEnabled
	}

	// Update the timers in memory
	// The divider register is simply the 8 most significant bits of the CPU
	// clock
	t.state.mmu.setIORAM(dividerAddr, uint8(t.cpuClock>>8))
	t.state.mmu.setIORAM(timaAddr, tima)
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

// onTIMAWrite is called when the TIMA register is written to. This protects
// the TIMA from being written to if it was recently updated to the TMA
// register's value.
func (t *timers) onTIMAWrite(addr uint16, writeVal uint8) uint8 {
	if t.tmaToTIMATransferCountdown > 0 {
		// The TIMA is protected from writes
		tima := t.state.mmu.atIORAM(timaAddr)
		return tima
	}
	return writeVal
}

// onTMAWrite is called when the TMA register is written to. This emulates a
// special behavior with the TMA->TIMA loading process. If instructions write
// to this address while the TMA->TIMA transfer is happening, the TIMA will
// take on this new value.
func (t *timers) onTMAWrite(addr uint16, writeVal uint8) uint8 {
	if t.tmaToTIMATransferCountdown > 0 {
		// During this time where the TIMA is being set to the TAC, any changes
		// to the TAC will also be reflected in the TIMA
		t.state.mmu.setIORAM(timaAddr, writeVal)
	}

	return writeVal
}
