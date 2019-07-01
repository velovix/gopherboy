package gameboy

import "fmt"

const (
	// The Game Boy processor clock speed
	cpuClockRate = 4194304

	// The number of CPU clock cycles in one machine cycle, or m-cycle.
	ticksPerMCycle = 4
)

// timers keeps track of all timers in the Gameboy, including the TIMA.
type timers struct {
	// A clock that increments every M-Cycle
	cpuClock uint16

	// Conceptually, the TIMA is a clock that runs at some frequency. In
	// reality, however, the TIMA is a variable that increments when a falling
	// edge detector detects a falling edge. The input to this falling edge
	// detector is (timaRunning & timaBit).
	//
	// The "delay" value of the falling edge detector used to increment the
	// TIMA. This holds the value previously fed to the falling edge detector.
	// If this value is 1 and the new value is 0, we know there's a falling
	// edge.
	//
	// The term "delay" wouldn't be my first choice to describe this concept,
	// but it's what they call it in most diagrams for falling edge detectors.
	fallingEdgeDetectorDelay uint8

	// True if the TIMA is overflowing to zero during this M-Cycle.
	timaOverflowing bool
	// True if the TMA is being transferred to the TIMA this M-Cycle. Note that
	// this happens one instruction after the TIMA overflows.
	tmaToTIMATransferring bool

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

	t.cpuClock++
	t.divider = uint8(t.cpuClock >> 6)

	if t.tmaToTIMATransferring {
		// This process was finished last M-Cycle
		t.tmaToTIMATransferring = false
	}

	// Check the countdown to see if it's time to process a delayed TIMA
	// interrupt
	if t.timaOverflowing {
		// Transfer the TMA value to the TIMA
		t.tima = t.tma
		t.timaOverflowing = false
		t.tmaToTIMATransferring = true

		if t.state.interruptsEnabled && t.interruptManager.timaEnabled() {
			// Flag a TIMA overflow interrupt
			t.interruptManager.flagTIMA()
		}
	}

	// Check for a falling edge and increment the TIMA if there was one
	timaBit := t.timaBit()
	fallingEdgeDetectorInput := uint8(0)
	if timaRunning && timaBit == 1 {
		fallingEdgeDetectorInput = 1
	}
	if fallingEdgeDetectorInput == 0 && t.fallingEdgeDetectorDelay == 1 {
		t.incrementTIMA()
	}

	// Update the delay value
	t.fallingEdgeDetectorDelay = fallingEdgeDetectorInput
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

	// Update the TAC value
	t.tac = writeVal

	timaRunningAfter := t.tac&0x4 == 0x4
	newTIMABit := t.timaBit()

	// Setting the TAC can bring the falling edge detector's input value low by
	// either disabling the TIMA or moving the TIMA bit to a bit that is low.
	// If the "delay" value of the falling edge detector was high, this will
	// trigger a falling edge and increment the TIMA.
	if t.fallingEdgeDetectorDelay == 1 &&
		(!timaRunningAfter || newTIMABit == 0) {

		t.incrementTIMA()
	}

	// All unused bits are high
	return 0xF8 | writeVal
}

// onTIMAWrite is called when the TIMA register is written to. This protects
// the TIMA from being written to if it was recently updated to the TMA
// register's value.
func (t *timers) onTIMAWrite(addr uint16, writeVal uint8) uint8 {
	if t.timaOverflowing {
		// This write cancels the TIMA overflow detection, so the TMA->TIMA
		// transfer and the overflow interrupt will not happen next M-Cycle
		t.timaOverflowing = false
		t.tima = writeVal

		return writeVal
	} else if t.tmaToTIMATransferring {
		// The TMA overwrites the value set here
		return t.tima
	} else {
		// Set the TIMA value normally
		t.tima = writeVal
		return writeVal
	}
}

// onTMAWrite is called when the TMA register is written to. This emulates a
// special behavior with the TMA->TIMA loading process. If instructions write
// to this address while the TMA->TIMA transfer is happening, the TIMA will
// take on this new value.
func (t *timers) onTMAWrite(addr uint16, writeVal uint8) uint8 {
	t.tma = writeVal

	if t.tmaToTIMATransferring {
		// During this time where the TIMA is being set to the TMA, any changes
		// to the TMA will also be reflected in the TIMA
		t.tima = writeVal
	}

	return writeVal
}

// timaBit returns the bit in the CPU clock that is used by the falling edge
// detector to decide if the TIMA needs to be incremented.
func (t *timers) timaBit() uint8 {
	timaRateBits := t.tac & 0x3

	// Pull the bit of interest from the CPU clock
	switch timaRateBits {
	case 0x0:
		// TIMA configured at 4096 Hz
		return uint8((t.cpuClock >> 7) & 0x1)
	case 0x3:
		// TIMA configured at 16384 Hz
		return uint8((t.cpuClock >> 5) & 0x1)
	case 0x2:
		// TIMA configured at 65536 Hz
		return uint8((t.cpuClock >> 3) & 0x1)
	case 0x1:
		// TIMA configured at 262144 Hz
		return uint8((t.cpuClock >> 1) & 0x1)
	default:
		panic(fmt.Sprintf("invalid TIMA rate %v", timaRateBits))
	}
}

func (t *timers) incrementTIMA() {
	t.tima++

	if t.tima == 0 {
		// There is a 1 M-Cycle delay between the TIMA overflow and the
		// interrupt and reset to TMA
		t.timaOverflowing = true
	}
}
