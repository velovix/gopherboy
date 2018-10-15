package main

type register interface {
	// set sets the register's value. If the register is too small for the
	// value, it will be truncated bit-wise.
	set(val uint16)
	// get returns the register's value.
	get() uint16
	// size returns the size in bits of the register.
	size() int
}

type register8Bit struct {
	val uint8
}

func (reg *register8Bit) set(val uint16) {
	reg.val = uint8(val)
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

func (reg *register16Bit) set(val uint16) {
	reg.val = val
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

func (reg *registerCombined) set(val uint16) {
	reg.first.set(uint16(uint8(val >> 8)))
	reg.second.set(uint16(uint8(val)))
}

func (reg *registerCombined) get() uint16 {
	return (reg.first.get() << 8) |
		uint16(uint8(reg.second.get()))
}

func (reg *registerCombined) size() int {
	return 16
}

type environment struct {
	regs map[registerType]register
	mem  []uint8
}

func newEnvironment() environment {
	env := environment{
		regs: make(map[registerType]register),
		mem:  make([]uint8, 0xFFFF),
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
	env.getReg(regAF).set(0x01)
	// Set the stack pointer to a high initial value
	env.getReg(regSP).set(0xFFFE)
	// I don't know why these values are set this way
	env.getReg(regF).set(0xB0)
	env.getReg(regBC).set(0x0013)
	env.getReg(regDE).set(0x00D8)
	env.getReg(regHL).set(0x014D)
	// TODO(velovix): Set a bunch of other memory addresses

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

func (env *environment) setZeroFlag(on bool) {
	mask := uint16(0x80)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

func (env *environment) setSubtractFlag(on bool) {
	mask := uint16(0x40)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

func (env *environment) setHalfCarryFlag(on bool) {
	mask := uint16(0x20)
	if on {
		env.regs[regF].set(env.regs[regF].get() | mask)
	} else {
		env.regs[regF].set(env.regs[regF].get() & ^mask)
	}
}

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

	regAF = "(AF)"
	regBC = "(BC)"
	regDE = "(DE)"
	regHL = "(HL)"

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
