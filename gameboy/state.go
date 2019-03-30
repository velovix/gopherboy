package gameboy

// State holds the entire state of the Game Boy.
type State struct {
	regA *normalRegister8
	regB *normalRegister8
	regC *normalRegister8
	regD *normalRegister8
	regE *normalRegister8
	regF *flagRegister8
	regH *normalRegister8
	regL *normalRegister8

	regAF *registerCombined
	regBC *registerCombined
	regDE *registerCombined
	regHL *registerCombined
	regSP *normalRegister16
	regPC *normalRegister16

	// The active memory management unit.
	mmu *mmu
	// If this value is >0, it is decremented after every operation. When this
	// timer decrements to 0, interrupts are enabled. This is used to emulate
	// the EI instruction's delayed effects.
	enableInterruptsTimer int
	// The master interrupt switch. If this is false, no interrupts will be
	// processed.
	interruptsEnabled bool
	// If true, the CPU is halted and no instructions will be run until an
	// interrupt occurs, which will set this value to false.
	halted bool
	// If true, the Game Boy is in "stop mode". This means that the CPU is
	// halted and the screen is turned white. This mode is exited when a button
	// is pressed.
	stopped bool

	// A program counter value pointing to the start of the current
	// instruction.
	instructionStart uint16
}

// NewState creates a new Game Boy state object with special memory addresses
// initialized in accordance with the Game Boy's start up sequence.
func NewState(mmu *mmu) *State {
	state := &State{
		mmu: mmu,
	}
	state.regA = &normalRegister8{0}
	state.regB = &normalRegister8{0}
	state.regC = &normalRegister8{0}
	state.regD = &normalRegister8{0}
	state.regE = &normalRegister8{0}
	state.regH = &normalRegister8{0}
	state.regL = &normalRegister8{0}
	state.regF = &flagRegister8{0}
	state.regAF = &registerCombined{
		first:  state.regA,
		second: state.regF}
	state.regBC = &registerCombined{
		first:  state.regB,
		second: state.regC}
	state.regDE = &registerCombined{
		first:  state.regD,
		second: state.regE}
	state.regHL = &registerCombined{
		first:  state.regH,
		second: state.regL}
	state.regSP = &normalRegister16{0}
	state.regPC = &normalRegister16{0}

	state.instructionStart = state.regPC.get()

	return state
}

// incrementPC increments the program counter by 1 and returns the value that
// was at its previous location.
func (state *State) incrementPC() uint8 {
	poppedVal := state.mmu.at(state.regPC.get())
	state.regPC.set(state.regPC.get() + 1)

	return poppedVal
}

// instructionDone signals to the state object that an instruction has
// finished. This moves the instructionStart value to the current position of
// the program counter.
func (state *State) instructionDone() {
	state.instructionStart = state.regPC.get()
}

// popFromStack reads a value from the current stack position and increments
// the stack pointer.
func (state *State) popFromStack() uint8 {
	val := state.mmu.at(state.regSP.get())

	state.regSP.set(state.regSP.get() + 1)

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
	state.regSP.set(state.regSP.get() - 1)

	state.mmu.set(state.regSP.get(), val)
}

// pushToStack16 pushes a 16-bit value to the stack, decrementing the stack
// pointer twice.
func (state *State) pushToStack16(val uint16) {
	lower, upper := split16(val)
	state.pushToStack(upper)
	state.pushToStack(lower)
}

// getZeroFlag returns the state of the zero bit in the flag register.
func (state *State) getZeroFlag() bool {
	mask := uint8(0x80)
	return state.regF.get()&mask == mask
}

// setZeroFlag sets the zero bit in the flag register to the given value.
func (state *State) setZeroFlag(on bool) {
	mask := uint8(0x80)
	if on {
		state.regF.set(state.regF.get() | mask)
	} else {
		state.regF.set(state.regF.get() & ^mask)
	}
}

// getSubtractFlag returns the state of the subtract bit in the flag register.
func (state *State) getSubtractFlag() bool {
	mask := uint8(0x40)
	return state.regF.get()&mask == mask
}

// setSubtractFlag sets the subtract bit in the flag register to the given
// value.
func (state *State) setSubtractFlag(on bool) {
	mask := uint8(0x40)
	if on {
		state.regF.set(state.regF.get() | mask)
	} else {
		state.regF.set(state.regF.get() & ^mask)
	}
}

// getHalfCarryFlag returns the state of the half carry bit in the flag
// register.
func (state *State) getHalfCarryFlag() bool {
	mask := uint8(0x20)
	return state.regF.get()&mask == mask
}

// setHalfCarryFlag sets the half carry bit in the flag register to the given
// value.
func (state *State) setHalfCarryFlag(on bool) {
	mask := uint8(0x20)
	if on {
		state.regF.set(state.regF.get() | mask)
	} else {
		state.regF.set(state.regF.get() & ^mask)
	}
}

// getCarryFlag returns the state of the carry bit in the flag register.
func (state *State) getCarryFlag() bool {
	mask := uint8(0x10)
	return state.regF.get()&mask == mask
}

// setCarryFlag sets the carry bit in the flag register to the given value.
func (state *State) setCarryFlag(on bool) {
	mask := uint8(0x10)
	if on {
		state.regF.set(state.regF.get() | mask)
	} else {
		state.regF.set(state.regF.get() & ^mask)
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
