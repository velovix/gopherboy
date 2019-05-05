package gameboy

// interruptManager receives interrupts and moves the program counter to
// interrupt handlers where appropriate.
type interruptManager struct {
	state  *State
	timers *timers

	// The value of the interrupt flag register, whose value indicates whether
	// or not certain interrupts are scheduled to happen.
	interruptFlags uint8
	// The value of the interrupt enable register, whose value indicates
	// whether or not certain interrupts are allowed to happen.
	interruptEnable uint8
}

func newInterruptManager(state *State, timers *timers) *interruptManager {
	mgr := &interruptManager{
		state:  state,
		timers: timers,
	}

	mgr.state.mmu.subscribeTo(ifAddr, mgr.onIFWrite)
	mgr.state.mmu.subscribeTo(ieAddr, mgr.onIEWrite)

	return mgr
}

// check checks for any interrupts that have happened, and triggers an
// interrupt handler by moving the program counter where appropriate.
func (mgr *interruptManager) check() {
	// The address of the handler to jump to
	var target uint16

	// Check each bit of the interrupt flag to see if an interrupt
	// happened, and each bit of the interrupt enable flag to check if
	// we should process it. Then, the flag is cleared.
	// Order here is important. Interrupts with lower target addresses
	// have priority over higher ones.
	if mgr.interruptEnable&mgr.interruptFlags&0x01 == 0x01 {
		// VBlank interrupt
		target = vblankInterruptTarget
		mgr.interruptFlags &= ^uint8(0x01)
	} else if mgr.interruptEnable&mgr.interruptFlags&0x02 == 0x02 {
		// LCDC interrupt
		target = lcdcInterruptTarget
		mgr.interruptFlags &= ^uint8(0x02)
	} else if mgr.interruptEnable&mgr.interruptFlags&0x04 == 0x04 {
		// TIMA overflow interrupt
		target = timaOverflowInterruptTarget
		mgr.interruptFlags &= ^uint8(0x04)
	} else if mgr.interruptEnable&mgr.interruptFlags&0x08 == 0x08 {
		// Serial interrupt
		target = serialInterruptTarget
		mgr.interruptFlags &= ^uint8(0x08)
	} else if mgr.interruptEnable&mgr.interruptFlags&0x10 == 0x10 {
		// P10-P13 interrupt
		target = p1Thru4InterruptTarget
		mgr.interruptFlags &= ^uint8(0x10)
	}

	if target != 0 {
		if mgr.state.interruptsEnabled {
			// Dispatch the interrupt
			mgr.state.interruptsEnabled = false
			// Push the current program counter to the stack for later use
			mgr.state.pushToStack16(mgr.state.regPC.get())
			mgr.state.regPC.set(target)
			// Dispatching an interrupt takes clock cycles
			for i := 0; i < interruptDispatchMCycles; i++ {
				mgr.timers.tick()
			}
		}

		if mgr.state.halted {
			// Take the Game Boy off halt mode. Note that this will happen even
			// if the master interrupt switch is disabled
			mgr.state.halted = false
			// Taking the Game Boy off halt mode takes one M-Cycle
			mgr.timers.tick()
		}
	}
}

// vblankEnabled returns true if the VBlank interrupt is enabled.
func (mgr *interruptManager) vblankEnabled() bool {
	return mgr.interruptEnable&0x1 == 0x1
}

// flagVBlank sets the VBlank interrupt flag, signaling that the interrupt
// should be run the next time interrupts are checked.
func (mgr *interruptManager) flagVBlank() {
	mgr.interruptFlags |= 0x1
}

// lcdcEnabled returns true if the LCDC interrupt is enabled.
func (mgr *interruptManager) lcdcEnabled() bool {
	return mgr.interruptEnable&0x2 == 0x2
}

// flagLCDC sets the LCDC interrupt flag, signaling that the interrupt should
// be run the next time interrupts are checked.
func (mgr *interruptManager) flagLCDC() {
	mgr.interruptFlags |= 0x2
}

// timaEnabled returns true if the TIMA interrupt is enabled.
func (mgr *interruptManager) timaEnabled() bool {
	return mgr.interruptEnable&0x4 == 0x4
}

// flagTIMA sets the TIMA interrupt flag, signaling that the interrupt should
// be run the next time interrupts are checked.
func (mgr *interruptManager) flagTIMA() {
	mgr.interruptFlags |= 0x4
}

// serialIOEnabled returns true if the serial IO interrupt is enabled.
func (mgr *interruptManager) serialIOEnabled() bool {
	return mgr.interruptEnable&0x8 == 0x8
}

// flagSerialIO sets the serial IO interrupt flag, signaling that the interrupt
// should be run the next time interrupts are checked.
func (mgr *interruptManager) flagSerialIO() {
	mgr.interruptFlags |= 0x8
}

// p10ToP13Enabled returns true if the P10-P13 interrupt is enabled.
func (mgr *interruptManager) p10ToP13Enabled() bool {
	return mgr.interruptEnable&0x10 == 0x10
}

// flagP10ToP13 sets the P10-P13 interrupt flag, signaling that the interrupt
// should be run the next time interrupts are checked.
// TODO(velovix): What a lame name. I need to find a better one.
func (mgr *interruptManager) flagP10ToP13() {
	mgr.interruptFlags |= 0x10
}

// onIFWrite is triggered when the Interrupt Flag register is written to. No
// interrupt handling action is taken here, but the unused bits in this
// register are set to 1.
func (mgr *interruptManager) onIFWrite(addr uint16, value uint8) uint8 {
	// The first three bits of the register are unused
	value |= 0xE3

	mgr.interruptFlags = value

	return value
}

// onIEWrite is triggered when the Interrupt Enable register is written to. It
// simply updates the internal value stored in this component.
func (mgr *interruptManager) onIEWrite(addr uint16, value uint8) uint8 {
	mgr.interruptEnable = value

	return value
}
