package gameboy

import (
	"time"
)

const (
	// The Game Boy processor clock speed
	cpuClockRate = 4194304

	// The number of CPU clock cycles in one machine cycle, or m-cycle.
	ticksPerMCycle = 4

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

	// The divider is a one-byte timer that is incremented every 64 clocks. It
	// is, in effect, the upper byte of the CPU clock, if we think of the
	// system clock as a two-byte value.
	divider uint8
	// The TIMA is a one-byte that can be configured to increment at various
	// rates. It is accessed as a memory register.
	tima uint8
	// The TAC is a one-byte area of one-bit flags that configure the TIMA.
	// TODO(velovix): Add some documentation here about what each bit does
	tac uint8
	// The TMA is a one-byte integer that configures the value the TIMA gets
	// set to when it overflows.
	tma uint8

	state            *State
	interruptManager *interruptManager
}

func newTimers(state *State) *timers {
	t := &timers{state: state}

	t.state.mmu.subscribeTo(dividerAddr, t.onDividerWrite)
	t.state.mmu.subscribeTo(tacAddr, t.onTACWrite)
	t.state.mmu.subscribeTo(timaAddr, t.onTIMAWrite)
	t.state.mmu.subscribeTo(tmaAddr, t.onTMAWrite)

	// The CPU is busy for 2 M-Cycles before running the boot ROM. The first
	// apparently sets up something related to the CPU's reset functionality.
	// The second pre-fetches the first instruction of the boot ROM. These
	// specifics are all internal details though so it's sufficient to simply
	// increment the timers.
	t.tick()
	t.tick()

	return t
}

// tick increments the timers by one m-cycle.
func (t *timers) tick() {
	// Parse the TAC bits for TIMA configuration information
	timaRunning := t.tac&0x4 == 0x4
	timaRateBits := t.tac & 0x3

	for i := 0; i < ticksPerMCycle; i++ {
		t.cpuClock++
		if t.cpuClock == cpuClockRate {
			t.cpuClock = 0
		}
		t.divider = uint8(t.cpuClock >> 8)

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
				t.tima = t.tma

				if t.state.interruptsEnabled && t.interruptManager.timaEnabled() {
					// Flag a TIMA overflow interrupt
					t.interruptManager.flagTIMA()
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
			t.tima++

			if t.tima == 0 {
				// There is a 4-cycle delay between the TIMA overflow and the
				// interrupt and reset to TMA
				t.timaInterruptCountdown = 4
			}
		}
		t.timaDelay = timaBitAndEnabled
	}
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

	t.tac = writeVal

	// All unused bits are high
	return 0xF8 | writeVal
}

// onTIMAWrite is called when the TIMA register is written to. This protects
// the TIMA from being written to if it was recently updated to the TMA
// register's value.
func (t *timers) onTIMAWrite(addr uint16, writeVal uint8) uint8 {
	if t.tmaToTIMATransferCountdown > 0 {
		// The TIMA is protected from writes
		return t.tima
	}
	t.tima = writeVal
	return writeVal
}

// onTMAWrite is called when the TMA register is written to. This emulates a
// special behavior with the TMA->TIMA loading process. If instructions write
// to this address while the TMA->TIMA transfer is happening, the TIMA will
// take on this new value.
func (t *timers) onTMAWrite(addr uint16, writeVal uint8) uint8 {
	t.tma = writeVal

	if t.tmaToTIMATransferCountdown > 0 {
		// During this time where the TIMA is being set to the TMA, any changes
		// to the TMA will also be reflected in the TIMA
		t.tima = writeVal
	}

	return writeVal
}
