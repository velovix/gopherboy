package gameboy

// interruptManager receives interrupts and moves the program counter to
// interrupt handlers where appropriate.
type interruptManager struct {
	state  *State
	timers *timers
}

func newInterruptManager(state *State, timers *timers) *interruptManager {
	mgr := &interruptManager{
		state:  state,
		timers: timers,
	}

	mgr.state.mmu.subscribeTo(ifAddr, mgr.onIFWrite)

	return mgr
}

// check checks for any interrupts that have happened, and triggers an
// interrupt handler by moving the program counter where appropriate.
func (mgr *interruptManager) check() {
	interruptFlags := mgr.state.mmu.memory[ifAddr]
	interruptEnable := mgr.state.mmu.memory[ieAddr]
	clearedInterruptFlags := interruptFlags

	// The address of the handler to jump to
	var target uint16

	// Check each bit of the interrupt flag to see if an interrupt
	// happened, and each bit of the interrupt enable flag to check if
	// we should process it. Then, the flag is cleared.
	// Order here is important. Interrupts with lower target addresses
	// have priority over higher ones.
	if interruptEnable&interruptFlags&0x01 == 0x01 {
		// VBlank interrupt
		target = vblankInterruptTarget
		clearedInterruptFlags &= ^uint8(0x01)
	} else if interruptEnable&interruptFlags&0x02 == 0x02 {
		// LCDC interrupt
		target = lcdcInterruptTarget
		clearedInterruptFlags &= ^uint8(0x02)
	} else if interruptEnable&interruptFlags&0x04 == 0x04 {
		// TIMA overflow interrupt
		target = timaOverflowInterruptTarget
		clearedInterruptFlags &= ^uint8(0x04)
	} else if interruptEnable&interruptFlags&0x08 == 0x08 {
		// Serial interrupt
		target = serialInterruptTarget
		clearedInterruptFlags &= ^uint8(0x08)
	} else if interruptEnable&interruptFlags&0x10 == 0x10 {
		// P10-P13 interrupt
		target = p1Thru4InterruptTarget
		clearedInterruptFlags &= ^uint8(0x10)
	}

	if target != 0 {
		if mgr.state.interruptsEnabled {
			// Dispatch the interrupt
			mgr.state.interruptsEnabled = false
			// Clear the interrupt flag
			mgr.state.mmu.memory[ifAddr] = clearedInterruptFlags
			// Push the current program counter to the stack for later use
			mgr.state.pushToStack16(mgr.state.regPC.get())
			mgr.state.regPC.set(target)
			// Dispatching an interrupt takes clock cycles
			mgr.timers.tick(interruptDispatchCycles)
		}

		if mgr.state.halted {
			// Take the Game Boy of halt mode. Note that this will happen even
			// if the master interrupt switch is disabled
			mgr.state.halted = false
			mgr.timers.tick(unhaltCycles)
		}
	}
}

// onIFWrite is triggered when the Interrupt Flag register is written to. No
// interrupt handling action is taken here, but the unused bits in this
// register are set to 1.
func (mgr *interruptManager) onIFWrite(addr uint16, value uint8) uint8 {
	// The first three bits of the register are unused
	return value | 0xE3
}
