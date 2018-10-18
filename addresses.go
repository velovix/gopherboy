package main

const (
	_ uint16 = 0x0
	// dividerAddr points to an 8-bit integer that is incremented every 64
	// clocks.
	dividerAddr = 0xFF04
	// timaAddr points to the TIMA, an 8-bit integer that can be configured to
	// increment at various rates.
	timaAddr = 0xFF05
	// tmaAddr points to the TMA, an 8-bit integer that configures what value
	// to set the TIMA when it overflows.
	tmaAddr = 0xFF06
	// tacAddr points to the TAC, a one byte area where bits can be flipped to
	// configure the TIMA.
	tacAddr = 0xFF07
	// ifAddr points to the interrupt flags, whose value indicates whether or
	// not an interrupt is happening.
	//
	// Bit 4: TODO(velovix): What is this?
	// Bit 3: Serial I/O transfer interrupts
	// Bit 2: Timer overflow interrupts
	// Bit 1: LCDC interrupts
	// Bit 0: V-blank interrupts
	ifAddr = 0xFF0F
	// ieAddr points to the interrupt enable flags, which can be flipped to
	// configure which interrupts are enabled. A value in 1 in a bit means
	// "enabled".
	//
	// Bit order is the same as the interrupt flag.
	ieAddr = 0xFFFF

	// vblankInterruptTarget points to the location that will be jumped to on a
	// vblank interrupt.
	vblankInterruptTarget = 0x0040
	// lcdcInterruptTarget points to the location that will be jumped to on an
	// LCDC interrupt.
	lcdcInterruptTarget = 0x0048
	// timaOverflowInterruptTarget points to the location that will be jumped
	// to on a TIMA overflow interrupt.
	timaOverflowInterruptTarget = 0x0050
	// serialInterruptTarget points to the location that will be jumped to on a
	// serial interrupt.
	serialInterruptTarget = 0x0058
	// p1Thru4InterruptTarget points to the location that will be jumped to
	// when a keypad is pressed.
	// TODO(velovix): Improve my understanding of this and come up with a
	// better name
	p1Thru4InterruptTarget = 0x0060
)
