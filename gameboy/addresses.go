package gameboy

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
	// Bit 7: Unused, always 1
	// Bit 6: Unused, always 1
	// Bit 5: Unused, always 1
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

	// nr10Addr points to the "Sound Mode 1 Sweep" register. This controls the
	// frequency sweep effect of the pulse A sound channel.
	//
	// TODO(velovix): A better job at documenting this
	//
	// Bit 7: Unused, always 1
	// Bits 6-4: Controls the length of the frequency sweep. The length is
	//           n/128Hz, where N is the value written here.
	// Bit 3: If 0, the frequency increases. If 1, the frequency decreases.
	// Bits 2-0: The number to shift the last frequency by in order to get the
	//           next frequency
	nr10Addr = 0xFF10
	// nr11Addr points to the "Sound Mode 1 Length and Wave Pattern Duty"
	// register. This controls the duty cycle of the pulse A square wave and
	// the length that the wave should play for.
	//
	// Bits 7-6: Wave pattern duty cycle.
	//   00: 12.5% duty cycle
	//   01: 25% duty cycle
	//   10: 50% duty cycle
	//   11: 75% duty cycle, sounds the same as 25%
	// Bits 5-0: The duration to play the sound for. The length is
	//           (64-t)*(1/256) seconds, wehre t is the value at these bits.
	nr11Addr = 0xFF11
	// nr12Addr points to the "Sound Mode 1 Envelope" register. This allows for
	// a volume sweep effect.
	//
	// Bits 7-4: The initial volume of the note, with 0 being no sound.
	// Bits 3: Controls whether the volume sweeps up or down. 0 is down, 1 is
	//         up.
	// Bits 2-0: Controls at what rate the volume is changed, from 0-7. If 0,
	//           the sweep is disabled. When set, the volume sweeps up or down
	//           one unit every n/64 seconds.
	nr12Addr = 0xFF12
	// nr13Addr points to the "Sound Mode 1 Frequency Lo" register. This
	// register sets the lower 8 bits of pulse A's frequency.
	nr13Addr = 0xFF13
	// nr14Addr points to the "Sound Mode 1 Frequency Hi" register. This
	// register has some bits to start pulse A and the upper three bits of the
	// frequency selection.
	//
	// Bit 7: When set, pulse A restarts
	// Bit 6: If 1, this voice will be turned off when its duration finishes.
	//        If 0, it will play indefinitely.
	// Bits 2-0: The most significant 3 bits of pulse A's frequency
	nr14Addr = 0xFF14

	// nr21Addr points to the "Sound Mode 2 Length and Wave Pattern Duty"
	// register. This controls the duty cycle of the pulse B square wave and
	// the length that the wave should play for.
	//
	// Bits 7-6: Wave pattern duty cycle.
	//   00: 12.5% duty cycle
	//   01: 25% duty cycle
	//   10: 50% duty cycle
	//   11: 75% duty cycle, sounds the same as 25%
	// Bits 5-0: The duration to play the sound for. The length is
	//           (64-t)*(1/256) seconds, wehre t is the value at these bits.
	nr21Addr = 0xFF16

	// nr22Addr points to the "Sound Mode 2 Envelope" register. This allows for
	// a volume sweep effect.
	//
	// Bits 7-4: The initial volume of the note, with 0 being no sound.
	// Bits 3: Controls whether the volume sweeps up or down. 0 is down, 1 is
	//         up.
	// Bits 2-0: Controls at what rate the volume is changed, from 0-7. If 0,
	//           the sweep is disabled. When set, the volume sweeps up or down
	//           one unit every n/64 seconds.
	nr22Addr = 0xFF17

	// nr13Addr points to the "Sound Mode 2 Frequency Lo" register. This
	// register sets the lower 8 bits of pulse B's frequency.
	nr23Addr = 0xFF18

	// nr24Addr points to the "Sound Mode 2 Frequency Hi" register. This
	// register has some bits to start pulse B and the upper three bits of the
	// frequency selection.
	//
	// Bit 7: When set, pulse B restarts
	// Bit 6: If 1, this voice will be turned off when its duration finishes.
	//        If 0, it will play indefinitely.
	// Bits 2-0: The most significant 3 bits of the pulse B's frequency
	nr24Addr = 0xFF19

	// nr30Addr points to the Sound Mode 3 On/Off Memory Register. It turns on
	// and off the wave channel.
	//
	// Bit 7: If 1, the wave channel is turned on.
	// Bits 6-0: Unused, always 1
	nr30Addr = 0xFF1A

	// nr31Addr is the "Sound Mode 3 Length" register. This register sets the
	// length at which the note will be played.
	//
	// Length in seconds = (256-nr31)*(1/2)
	nr31Addr = 0xFF1B

	// nr32Addr points to the Sound Mode 3 Select Output Level Memory Register.
	// This register controls what is effectively the volume of the wave voice.
	// The wave voice's volume is decreased by shifting the wave value right by
	// the specified amount.
	//
	// Bits 6-5: Select output level
	//   00: Shift 4 bits, effectively muting the wave voice
	//   01: Produce wave pattern data as-is
	//   10: Produce wave pattern data shifted once to the right
	//   11: Produce wave pattern data shifted twice to the right
	nr32Addr = 0xFF1C

	// nr33Addr points to the "Sound Mode 3 Frequency Lo" register. This
	// register sets the lower 8 bits of the wave channel's frequency.
	nr33Addr = 0xFF1D

	// nr34Addr points to the "Sound Mode 3 Frequency Hi" register. This
	// register has some bits to start the wave channel and the upper three
	// bits of the frequency selection.
	//
	// Bit 7: When set, the wave channel restarts
	// Bit 6: If 1, this voice will be turned off when its duration finishes.
	//        If 0, it will play indefinitely.
	// Bits 2-0: The most significant 3 bits of the wave channel's frequency
	nr34Addr = 0xFF1E

	// nr41Addr points to the Sound Mode 4 Sound Length Memory Register. It
	// controls the length that the noise channel plays for.
	//
	// Bits 7-6: Unused, always 1
	// Bits 5-0: Controls sound length. If this value is t1, then the sound
	//           will play for (64-t1)*(1/256) seconds.
	nr41Addr = 0xFF20

	// nr42Addr points to the Sound Mode 4 Envelope Register. It controls the
	// volume sweep for the noise voice.
	//
	// Bits 7-4: The initial volume of the note, with 0 being no sound.
	// Bit 3: Controls whether the volume sweeps up or down. 0 is down, 1 is
	//        up.
	// Bits 2-0: Controls at what rate the volume is changed, from 0-7. If 0,
	//           the sweep is disabled. When set, the volume sweeps up or down
	//           one unit every n/64 seconds.
	nr42Addr = 0xFF21

	// nr43Addr points to the Sound Mode 4 Polynomial Register. It controls
	// various aspects of how the random noise is generated.
	//
	// The noise generation is done using a linear feedback shift register,
	// which provides pseudo-random numbers. We will refer to this as the LFSR.
	//
	// Bits 7-4: Controls the frequency that the LFSR is shifted at. Values
	//           0b1110 and 0b1111 are invalid. TODO(velovix): What happens?
	//           (dividing ratio) * 1/2^n
	// Bit 3: Selects the LFSR's size. If 0, it is 15-bit. If 1, it is 7-bit.
	// Bits 2-0: The dividing ratio of frequencies.
	//           0: f * 1/2^3 * 2
	//           n: f * 1/2^3 * 1/n
	//           Where f=4194304 Hz
	nr43Addr = 0xFF22

	// nr44Addr points to the Sound Mode 4 Counter/Consecutive Initial Memory
	// Register.
	//
	// Bit 7: When set, the noise channel restarts
	// Bit 6: If 1, this voice will be turned off when its duration finishes.
	//        If 0, it will play indefinitely.
	// Bit 5-0: Unused, always 1
	nr44Addr = 0xFF23

	// nr51Addr points to the Selection of Sound Output Terminal. It's a set of
	// bits which turn on and off the sound of each voice for the left and
	// right audio channels. 1 is on, 0 is off.
	//
	// Bit 7: Sound mode 4 for left side on/off
	// Bit 6: Sound mode 3 for left side on/off
	// Bit 5: Sound mode 2 for left side on/off
	// Bit 4: Sound mode 1 for left side on/off
	// Bit 3: Sound mode 4 for right side on/off
	// Bit 2: Sound mode 3 for right side on/off
	// Bit 1: Sound mode 2 for right side on/off
	// Bit 0: Sound mode 1 for right side on/off
	nr51Addr = 0xFF25

	// nr52Addr points to the Sound On/Off Memory Register. It's a set of
	// on/off indicators for each channel and all sound.
	//
	// Bit 7: All sound on/off switch. 1 if on.
	// Bits 6-4: Unused, always 1
	// Bit 3: Sound 4 on/off switch. 1 if on. Read-only.
	// Bit 2: Sound 3 on/off switch. 1 if on. Read-only.
	// Bit 1: Sound 2 on/off switch. 1 if on. Read-only.
	// Bit 0: Sound 1 on/off switch. 1 if on. Read-only.
	nr52Addr = 0xFF26

	// Designates the area in RAM where the custom wave pattern is specified.
	wavePatternRAMStart = 0xFF30
	wavePatternRAMEnd   = 0xFF40

	// scrollYAddr points to the "Scroll Y" memory register. This controls the
	// position of the top left of the background.
	scrollYAddr = 0xFF42
	// scrollXAddr points to the "Scroll X" memory register. This controls the
	// position of the top left of the background.
	scrollXAddr = 0xFF43
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

	// key1Addr has something to do with switching to double-speed mode on the
	// CGB. Since this is a DMG emulator for the time being, this is unused.
	key1Addr = 0xFF4D

	// vbkAddr is a VRAM bank register. It's a CGB thing that this emulator
	// doesn't have to worry about yet
	vbkAddr = 0xFF4F

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
	// Bit 7: Unused, always 1
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

	// bootROMDisableAddr is the address that, when written to, disables the
	// boot ROM. The boot ROM itself writes to this when it is finished.
	bootROMDisableAddr = 0xFF50

	// These addresses have something to do with a "new" DMA transfer mode
	// introduced in the CGB. We don't have to worry about this since this is a
	// DMG emulator.
	hdma1Addr = 0xFF51
	hdma2Addr = 0xFF52
	hdma3Addr = 0xFF53
	hdma4Addr = 0xFF54
	hdma5Addr = 0xFF55

	// svbkAddr is a WRAM bank controller thing for the CGB. Unused on the DMG.
	svbkAddr = 0xFF70

	// rpAddr is some register for controlling infrared communications. Only
	// the CGB has infrared so this goes unused on the DMG.
	rpAddr = 0xFF56

	// These addresses have something to do with the CGB color palette. They go
	// unused on the DMG.
	bcpsAddr = 0xFF68
	bcpdAddr = 0xFF69
	ocpsAddr = 0xFF6A
	ocpdAddr = 0xFF6B

	// These addresses probably have something to do with sound on the CGB.
	// They're unused on the DMG.
	pcm12Ch2Addr = 0xFF76
	pcm34Ch4Addr = 0xFF77

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
	// Bit 7: Unused, always 1
	// Bit 6: Unused, always 1
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

	// sbAddr points to the Serial Transfer Data Memory Register, which is a
	// place where 8 bits of serial data is read from or written to.
	sbAddr = 0xFF01

	// scAddr points to the SIO Control Memory Register, which allows for
	// serial communication control.
	// TODO(velovix): Fill this out more as I learn more about Game Boy serial.
	//
	// Bit 7: If set to 1, a transfer will be initiated
	// Bit 6-1: Unused, always 1
	// Bit 0: Shift clock TODO(velovix): What is this?
	//        0: External clock
	//        1: Internal clock
	scAddr = 0xFF02

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
