package main

// State holds the entire state of the Game Boy.
type State struct {
	// A map of 8-bit register names to their corresponding register.
	regs8 map[registerType]register8
	// A map of 16-bit register names to their corresponding register.
	regs16 map[registerType]register16
	// The active memory management unit.
	mmu *mmu
	// If this value is >0, it is decremented after every operation. When this
	// timer decrements to 0, interrupts are enabled. This is used to emulate
	// the EI instruction's delayed effects.
	enableInterruptsTimer int
	// If this value is >0, it is decremented after every operation. When this
	// timer decrements to 0, interrupts are disabled. This is used to emulate
	// the DI instruction's delayed effects.
	disableInterruptsTimer int
	// The master interrupt switch. If this is false, no interrupts will be
	// processed.
	interruptsEnabled bool
	// If true, the processor will not run instructions until an interrupt
	// occurs.
	waitingForInterrupts bool
	// If true, the Game Boy is in "stop mode". This means that the CPU is
	// halted and the screen is turned white. This mode is exited when a button
	// is pressed.
	stopped bool
}

// NewState creates a new Game Boy state object with special memory addresses
// initialized in accordance with the Game Boy's start up sequence.
func NewState(mmu *mmu) *State {
	state := &State{
		regs8:  make(map[registerType]register8),
		regs16: make(map[registerType]register16),
		mmu:    mmu,
	}

	state.regs8[regA] = &normalRegister8{0}
	state.regs8[regB] = &normalRegister8{0}
	state.regs8[regC] = &normalRegister8{0}
	state.regs8[regD] = &normalRegister8{0}
	state.regs8[regE] = &normalRegister8{0}
	state.regs8[regH] = &normalRegister8{0}
	state.regs8[regL] = &normalRegister8{0}

	state.regs8[regF] = &flagRegister8{0}

	state.regs16[regAF] = &registerCombined{
		first:  state.regs8[regA],
		second: state.regs8[regF]}
	state.regs16[regBC] = &registerCombined{
		first:  state.regs8[regB],
		second: state.regs8[regC]}
	state.regs16[regDE] = &registerCombined{
		first:  state.regs8[regD],
		second: state.regs8[regE]}
	state.regs16[regHL] = &registerCombined{
		first:  state.regs8[regH],
		second: state.regs8[regL]}

	state.regs16[regSP] = &normalRegister16{0}
	state.regs16[regPC] = &normalRegister16{0}

	// Set registers to their initial value
	// 0x100 is the designated entry point of a Gameboy ROM
	state.regs16[regPC].set(0x100)
	// Set the stack pointer to a high initial value
	state.regs16[regSP].set(0xFFFE)
	// I don't know why these values are set this way
	state.regs16[regAF].set(0x01B0)
	state.regs16[regBC].set(0x0013)
	state.regs16[regDE].set(0x00D8)
	state.regs16[regHL].set(0x014D)

	// Set memory addresses
	state.mmu.setNoNotify(timaAddr, 0x00)
	state.mmu.setNoNotify(tmaAddr, 0x00)
	state.mmu.setNoNotify(tacAddr, 0x00)
	state.mmu.setNoNotify(ieAddr, 0x00)
	state.mmu.setNoNotify(lcdcAddr, 0x91)
	state.mmu.setNoNotify(scrollYAddr, 0x00)
	state.mmu.setNoNotify(scrollXAddr, 0x00)
	state.mmu.setNoNotify(lycAddr, 0x00)
	state.mmu.setNoNotify(bgpAddr, 0xFC)
	state.mmu.setNoNotify(opb0Addr, 0xFF)
	state.mmu.setNoNotify(opb1Addr, 0xFF)
	state.mmu.setNoNotify(windowPosYAddr, 0x00)
	state.mmu.setNoNotify(windowPosXAddr, 0x00)
	state.mmu.setNoNotify(ieAddr, 0x00)
	// TODO(velovix): Set even more memory addresses

	return state
}

// incrementPC increments the program counter by 1 and returns the value that
// was at its previous location.
func (state *State) incrementPC() uint8 {
	poppedVal := state.mmu.at(state.regs16[regPC].get())
	state.regs16[regPC].set(state.regs16[regPC].get() + 1)

	return poppedVal
}

// popFromStack reads a value from the current stack position and increments
// the stack pointer.
func (state *State) popFromStack() uint8 {
	val := state.mmu.at(state.regs16[regSP].get())

	state.regs16[regSP].set(state.regs16[regSP].get() + 1)

	return val
}

// popFromStack16 reads a 16-bit value from the current stack position and
// decrements the stack pointer twice.
func (state *State) popFromStack16() uint16 {
	lower := state.popFromStack()
	upper := state.popFromStack()

	return combine16(lower, upper)
}

// pushToStack decrements the stack pointer and writes the given value.
func (state *State) pushToStack(val uint8) {
	state.regs16[regSP].set(state.regs16[regSP].get() - 1)

	state.mmu.set(state.regs16[regSP].get(), val)
}

// pushToStack16 pushes a 16-bit value to the stack, decrementing the stack
// pointer twice.
func (state *State) pushToStack16(val uint16) {
	lower, upper := split16(val)
	state.pushToStack(upper)
	state.pushToStack(lower)
}

// relativeJump moves the program counter by the given signed value.
func (state *State) relativeJump(offset int) {
	state.regs16[regPC].set(uint16(int(state.regs16[regPC].get()) + offset))
}

// getZeroFlag returns the state of the zero bit in the flag register.
func (state *State) getZeroFlag() bool {
	mask := uint8(0x80)
	return state.regs8[regF].get()&mask == mask
}

// setZeroFlag sets the zero bit in the flag register to the given value.
func (state *State) setZeroFlag(on bool) {
	mask := uint8(0x80)
	if on {
		state.regs8[regF].set(state.regs8[regF].get() | mask)
	} else {
		state.regs8[regF].set(state.regs8[regF].get() & ^mask)
	}
}

// getSubtractFlag returns the state of the subtract bit in the flag register.
func (state *State) getSubtractFlag() bool {
	mask := uint8(0x40)
	return state.regs8[regF].get()&mask == mask
}

// setSubtractFlag sets the subtract bit in the flag register to the given
// value.
func (state *State) setSubtractFlag(on bool) {
	mask := uint8(0x40)
	if on {
		state.regs8[regF].set(state.regs8[regF].get() | mask)
	} else {
		state.regs8[regF].set(state.regs8[regF].get() & ^mask)
	}
}

// getHalfCarryFlag returns the state of the half carry bit in the flag
// register.
func (state *State) getHalfCarryFlag() bool {
	mask := uint8(0x20)
	return state.regs8[regF].get()&mask == mask
}

// setHalfCarryFlag sets the half carry bit in the flag register to the given
// value.
func (state *State) setHalfCarryFlag(on bool) {
	mask := uint8(0x20)
	if on {
		state.regs8[regF].set(state.regs8[regF].get() | mask)
	} else {
		state.regs8[regF].set(state.regs8[regF].get() & ^mask)
	}
}

// getCarryFlag returns the state of the carry bit in the flag register.
func (state *State) getCarryFlag() bool {
	mask := uint8(0x10)
	return state.regs8[regF].get()&mask == mask
}

// setCarryFlag sets the carry bit in the flag register to the given value.
func (state *State) setCarryFlag(on bool) {
	mask := uint8(0x10)
	if on {
		state.regs8[regF].set(state.regs8[regF].get() | mask)
	} else {
		state.regs8[regF].set(state.regs8[regF].get() & ^mask)
	}
}

// flagMask is a bit mask that can be applied to the flag register to check the
// flag's value.
const (
	_ uint8 = 0x00
	// zeroFlag is set when the result of the previous math operation was zero.
	zeroFlag = 0x80
	// subtractFlag is set when the previous math operation was a subtraction.
	subtractFlag = 0x40
	// halfCarryFlag is set when the previous math operation results in a carry
	// to the 4th bit.
	halfCarryFlag = 0x20
	// carryFlag is set when the previous math operation results in a carry
	// from the most significant bit.
	carryFlag = 0x10
)
