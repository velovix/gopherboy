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
	// Bit 4: P10-P13 high-to-low interrupts. Happens when a button is pressed.
	// Bit 3: Serial I/O transfer interrupts
	// Bit 2: TIMA overflow interrupts. Happens when the 8-bit configurable
	//        timer overflows.
	// Bit 1: LCDC interrupts. Happens when certain events happen on the
	//        display. What exactly triggers this interrupt can be configured
	//        using the LCDC memory register.
	// Bit 0: V-blank interrupts. Happens when the V-blank period starts on the
	//        display.
	ifAddr = 0xFF0F
	// ieAddr points to the interrupt enable flags, which can be flipped to
	// configure which interrupts are enabled. A value in 1 in a bit means
	// "enabled".
	//
	// Bit order is the same as the interrupt flag.
	ieAddr = 0xFFFF
	// windowPosYAddr points to the "Window Position Y" memory register. This
	// controls the position of the window in the Y direction.
	windowPosYAddr = 0xFF4A
	// windowPosXAddr points to the "Window Position X" memory register. This
	// controls the position of the window in the X direction.
	windowPosXAddr = 0xFF4B
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
	// bgpAddr points to the BGP memory register, which controls the
	// background and window palette. This register maps dot data to actual
	// colors.
	//
	// Colors go from 0b11 to 0b00, where 0b11 is the darkest and 0b00 is the
	// lightest.
	//
	// Bits 7-6: The color for dot data 0b11
	// Bits 5-4: The color for dot data 0b10
	// Bits 3-2: The color for dot data 0b01
	// Bits 1-0: The color for dot data 0b00
	bgpAddr = 0xFF47
	// lyAddr points to the LCD Current Scanline memory register, which
	// indicates what line of the display is currently being drawn. Starts at 0
	// and ends at 153.
	lyAddr = 0xFF44
	// lycAddr points to the LY Compare memory register. Users can write a
	// value to this register. When the LY memory register value is the same as
	// this register's value, an interrupt can be generated.
	lycAddr = 0xFF45
	// statAddr points to the STAT memory register. This register is used to
	// check the status of the LCD and to configure LCD-related interrupts.
	//
	// Bit 7: Unused, always 1 (read-only)
	// Bit 6: If 1, an interrupt will be triggered when the LY register equals
	//        the LYC register
	// Bit 5: If 1, the LCDC interrupt will be triggered when the video
	//        controller enters mode 2.
	// Bit 4: If 1, the LCDC interrupt will be triggered when the video
	//        controller enters mode 1.
	// Bit 3: If 1, the LCDC interrupt will be triggered when the video
	//        controller enters mode 0.
	// Bit 2: Set to 1 if the LY memory register value is the same as the LYC
	//        memory register value. (read-only)
	// Bit 1-0: The current mode as a 2-bit number.
	//     Mode 0: We're in an HBlank period
	//     Mode 1: We're in a VBlank period
	//     Mode 2: Controller is reading from OAM RAM. OAM cannot be written
	//             to.
	//     Mode 3: Controller is transferring data from OAM and VRAM. OAM and
	//             VRAM cannot be written to.
	statAddr = 0xFF41

	// videoRAMStart is the address where video RAM starts in memory.
	videoRAMStart = 0x8000
	// videoRAMEnd is the address where video RAM stops in memory.
	videoRAMEnd = 0xA000

	// Memory addresses for the two available tile data tables.
	// Note that this data table actually starts at 0x8800, but tile values
	// that reference this table can be negative, allowing them to access the
	// data before this address.
	tileDataTable0 = 0x9000
	tileDataTable1 = 0x8000

	// Memory address for sprite data.
	spriteDataTable = 0x8000

	// Memory addresses for the two available tile maps.
	tileMap0 = 0x9800
	tileMap1 = 0x9C00

	// p1Addr points to the Joypad Memory Register, which is used to query
	// button/joypad input.
	//
	// Bit 7: Unused
	// Bit 6: Unused
	// Bit 5: If set to 0, bits 3-0 will provide the status of the buttons (A,
	//        B, Select, Start)
	// Bit 4: If set to 0, bits 3-0 will provide the status of the d-pad (Up,
	//        Down, Left, Right)
	// Bit 3: Represents the state of the down button on the d-pad or the start
	//        button, depending on the values of bits 4 and 3. 0=pressed.
	// Bit 2: Represents the state of the up button on the d-pad or the select
	//        button, depending on the values of bits 4 and 3. 0=pressed.
	// Bit 1: Represents the state of the left button on the d-pad or the B
	//        button, depending on the values of bits 4 and 3. 0=pressed.
	// Bit 0: Represents the state of the right button on the d-pad or the A
	//        button, depending on the values of bits 4 and 3. 0=pressed.
	p1Addr = 0xFF00

	// These variables points to the OBP0 and OPB1 memory registers, which
	// control the two available sprite palettes. These registers maps dot data
	// to actual colors.
	//
	// Colors go from 0b11 to 0b00, where 0b11 is the darkest and 0b00 is the
	// lightest.
	//
	// Bits 7-6: The color for dot data 0b11
	// Bits 5-4: The color for dot data 0b10
	// Bits 3-2: The color for dot data 0b01
	// Bits 1-0: Unused. This dot data is always interpreted as transparent.
	opb0Addr = 0xFF48
	opb1Addr = 0xFF48

	// dmaAddr points to the DMA Transfer and Start Address register. When this
	// register is written to, a transfer will happen between a specified
	// memory address and OAM RAM. Games use this to update sprite information
	// automatically.
	//
	// The transfer uses the written value as the upper byte of the source
	// address for OAM data. For example, if a game writes a 0x28 to the DMA,
	// a transfer will begin from 0x2800-0x289F to 0xFE00-0xFE9F (OAM RAM).
	dmaAddr = 0xFF46

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
	p1Thru4InterruptTarget = 0x0060
)
