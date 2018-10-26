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
	// scrollYAddr points to the "Scroll Y" memory register. This controls the
	// position of the top left of the background.
	scrollYAddr = 0xFF42
	// scrollXAddr points to the "Scroll X" memory register. This controls the
	// position of the top left of the background.
	scrollXAddr = 0xFF43
	// ifAddr points to the interrupt flags, whose value indicates whether or
	// not an interrupt is happening.
	//
	// Bit 4: TODO(velovix): What is this?
	// Bit 3: Serial I/O transfer interrupts
	// Bit 2: TIMA overflow interrupts
	// Bit 1: LCDC interrupts
	// Bit 0: V-blank interrupts
	ifAddr = 0xFF0F
	// ieAddr points to the interrupt enable flags, which can be flipped to
	// configure which interrupts are enabled. A value in 1 in a bit means
	// "enabled".
	//
	// Bit order is the same as the interrupt flag.
	ieAddr = 0xFFFF
	// wndPosYAddr points to the "Window Position Y" memory register. This
	// controls the position of the window in the Y direction.
	wndPosYAddr = 0xFF4A
	// wndPosXAddr points to the "Window Position X" memory register. This
	// controls the position of the window in the X direction.
	wndPosXAddr = 0xFF4B
	// lcdcAddr points to the LCDC memory register, which controls various
	// aspects of how a frame is drawn.
	//
	// Bit 7: Turns the LCD on and off.
	//   0: LCD is off
	//   1: LCD is on
	// Bit 6: Selects which tile map should be used for the window.
	//   0: 0x9800-0x9BFF
	//   1: 0x9C00-0x9FFF
	// Bit 5: Turns the window on and off
	//   0: Window is off
	//   1: Window is on
	// Bit 4: Selects which tile data table the window and background use.
	//   0: 0x8800-0x97FF
	//   1: 0x8000-0x8FFF
	// Bit 3: Selects which tile map should be used for the background.
	//   0: 0x9800-0x9BFF
	//   1: 0x9C00-0x9FFF
	// Bit 2: Selects which sprite size we're currently using
	//   0: 8x8 sprites
	//   1: 8x16 sprites
	// Bit 1: Turns sprites on and off
	//   0: Sprites are off
	//   1: Sprites are on
	// Bit 0: Turns the background and window on and off
	//   0: Both are off
	//   1: Both are on
	lcdcAddr = 0xFF40

	// This is the last address of ROM bank 0, a bank that always contains the
	// first 0x3FFF bytes of cartridge data.
	romBank0End = 0x4000
	// videoRAMStart is the address where video RAM starts in memory.
	videoRAMStart = 0x8000
	// videoRAMEnd is the address where video RAM stops in memory.
	videoRAMEnd = 0xA000

	// Memory addresses for the two available tile data tables.
	// Note that this data table actually starts at 0x8800, but tile values
	// that reference this table can be negative, allowing them to access the
	// data before this address.
	tileDataTable0 = 0x8800
	tileDataTable1 = 0x8000

	// Memory addresses for the two available tile maps.
	tileMap0 = 0x9800
	tileMap1 = 0x9C00

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
