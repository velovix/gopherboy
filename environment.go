package main

type register interface {
	// set sets the register's value. If the register is too small for the
	// value, it will be truncated bit-wise. Returns the set result.
	set(val uint16) uint16
	// get returns the register's value.
	get() uint16
	// size returns the size in bits of the register.
	size() int
}

type register8Bit struct {
	val uint8
}

func (reg *register8Bit) set(val uint16) uint16 {
	reg.val = uint8(val)
	return reg.get()
}

func (reg *register8Bit) get() uint16 {
	return uint16(reg.val)
}

func (reg *register8Bit) size() int {
	return 8
}

type register16Bit struct {
	val uint16
}

func (reg *register16Bit) set(val uint16) uint16 {
	reg.val = val
	return reg.get()
}

func (reg *register16Bit) get() uint16 {
	return reg.val
}

func (reg *register16Bit) size() int {
	return 16
}

type registerCombined struct {
	first, second register
}

func (reg *registerCombined) set(val uint16) uint16 {
	reg.first.set(uint16(uint8(val >> 8)))
	reg.second.set(uint16(uint8(val)))

	return reg.get()
}

func (reg *registerCombined) get() uint16 {
	return (reg.first.get() << 8) |
		uint16(uint8(reg.second.get()))
}

func (reg *registerCombined) size() int {
	return 16
}

type environment struct {
	// A map of register names to their corresponding register.
	regs map[registerType]register
	// All memory in the Gameboy
	mem []uint8
	// The master interrupt switch. If this is false, no interrupts will be
	// processed.
	interruptsEnabled bool
	// If true, the processor will not run instructions until an interrupt
	// occurs.
	waitingForInterrupts bool
}

func newEnvironment() *environment {
	env := &environment{
		regs: make(map[registerType]register),
		mem:  make([]uint8, 0x10000),
	}

	env.regs[regA] = &register8Bit{0}
	env.regs[regB] = &register8Bit{0}
	env.regs[regC] = &register8Bit{0}
	env.regs[regD] = &register8Bit{0}
	env.regs[regE] = &register8Bit{0}
	env.regs[regH] = &register8Bit{0}
	env.regs[regL] = &register8Bit{0}

	env.regs[regF] = &register8Bit{0}

	env.regs[regAF] = &registerCombined{
		first:  env.regs[regA],
		second: env.regs[regF]}
	env.regs[regBC] = &registerCombined{
		first:  env.regs[regB],
		second: env.regs[regC]}
	env.regs[regDE] = &registerCombined{
		first:  env.regs[regD],
		second: env.regs[regE]}
	env.regs[regHL] = &registerCombined{
		first:  env.regs[regH],
		second: env.regs[regL]}

	env.regs[regSP] = &register16Bit{0}
	env.regs[regPC] = &register16Bit{0}

	// Set registers to their initial value
	// This value depends on the system being emulated.
	// GB/SGB: 0x01 | GBP: 0xFF | CGB: 0x11
	env.regs[regAF].set(0x01)
	// Set the stack pointer to a high initial value
	env.regs[regSP].set(0xFFFE)
	// I don't know why these values are set this way
	env.regs[regF].set(0xB0)
	env.regs[regBC].set(0x0013)
	env.regs[regDE].set(0x00D8)
	env.regs[regHL].set(0x014D)

	// Set memory addresses
	// Set the timer values
	env.mem[timaAddr] = 0x00
	env.mem[tmaAddr] = 0x00
	env.mem[tacAddr] = 0x00
	env.mem[ieAddr] = 0x00
	// TODO(velovix): Set even more memory addresses

	return env
}

func (env *environment) getReg(reg registerType) register {
	return env.regs[reg]
}

// incrementPC increments the program counter by 1 and returns the value that
// was at its previous location.
func (env *environment) incrementPC() uint8 {
	poppedVal := env.mem[env.regs[regPC].get()]
	env.regs[regPC].set(env.regs[regPC].get() + 1)

	return poppedVal
}

// popFromStack reads a value from the current stack position and increments
// the stack pointer.
func (env *environment) popFromStack() uint8 {
	val := env.mem[env.regs[regSP].get()]

	env.regs[regSP].set(env.regs[regSP].get() + 1)

	return val
}

// popFromStack16 reads a 16-bit value from the current stack position and
// decrements the stack pointer twice.
func (env *environment) popFromStack16() uint16 {
	upper := env.popFromStack()
	lower := env.popFromStack()

	return combine(lower, upper)
}

// pushToStack decrements the stack pointer and writes the given value.
func (env *environment) pushToStack(val uint8) {
	env.regs[regSP].set(env.regs[regSP].get() - 1)

	env.mem[env.regs[regSP].get()] = val
}

// pushToStack16 pushes a 16-bit value to the stack, decrementing the stack
// pointer twice.
func (env *environment) pushToStack16(val uint16) {
	lower, upper := split16(val)
	env.pushToStack(lower)
	env.pushToStack(upper)
}

// relativeJump moves the program counter by the given signed value, including
// some special rules relative jumps have.
func (env *environment) relativeJump(offset int) {
	// Special case where backwards jumps always move back one more than the
	// given value
	if offset < 0 {
		offset--
	}
	env.regs[regPC].set(uint16(int(env.regs[regPC].get()) + offset))
}

// getZeroFlag returns the state of the zero bit in the flag register.
func (env *environment) getZeroFlag() bool {
	mask := uint16(0x80)
	return env.regs[regF].get()&mask == mask
}

// setZeroFlag sets the zero bit in the flag register to the given value.
func (env *environment) setZeroFlag(on bool) {
	mask := uint16(0x80)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

// getSubtractFlag returns the state of the subtract bit in the flag register.
func (env *environment) getSubtractFlag() bool {
	mask := uint16(0x40)
	return env.regs[regF].get()&mask == mask
}

// setSubtractFlag sets the subtract bit in the flag register to the given
// value.
func (env *environment) setSubtractFlag(on bool) {
	mask := uint16(0x40)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

// getHalfCarryFlag returns the state of the half carry bit in the flag
// register.
func (env *environment) getHalfCarryFlag() bool {
	mask := uint16(0x20)
	return env.regs[regF].get()&mask == mask
}

// setHalfCarryFlag sets the half carry bit in the flag register to the given
// value.
func (env *environment) setHalfCarryFlag(on bool) {
	mask := uint16(0x20)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

// getCarryFlag returns the state of the carry bit in the flag register.
func (env *environment) getCarryFlag() bool {
	mask := uint16(0x10)
	return env.regs[regF].get()&mask == mask
}

// setCarryFlag sets the carry bit in the flag register to the given value.
func (env *environment) setCarryFlag(on bool) {
	mask := uint16(0x10)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

type registerType string

const (
	_    registerType = ""
	regA              = "A"
	regB              = "B"
	regC              = "C"
	regD              = "D"
	regE              = "E"
	regH              = "H"
	regL              = "L"

	regAF = "AF"
	regBC = "BC"
	regDE = "DE"
	regHL = "HL"

	regSP = "SP"
	regF  = "F"
	regPC = "PC"
)

// flagMask is a bit mask that can be applied to the flag register to check the
// flag's value.
const (
	_ uint16 = 0x00
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
