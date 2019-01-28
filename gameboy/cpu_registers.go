package gameboy

// register8 represents an 8-bit register. Registers A, B, C, D, E, H, L, and F
// are all of this type.
type register8 interface {
	set(val uint8) uint8
	get() uint8
}

// normalRegister8 is a regular 8-bit register. It stores 8-bit values with no
// other special behavior.
type normalRegister8 struct {
	val uint8
}

func (reg *normalRegister8) set(val uint8) uint8 {
	reg.val = val
	return reg.get()
}

func (reg *normalRegister8) get() uint8 {
	return reg.val
}

// flagRegister8 is the 8-bit flag register. It is distinct from a normal 8-bit
// register because bits 3 through 0 are always zero. Any attempts to write to
// these bits will be ignored
type flagRegister8 struct {
	val uint8
}

func (reg *flagRegister8) set(val uint8) uint8 {
	// Mask away unused bits
	val &= 0xF0
	reg.val = val
	return reg.get()
}

func (reg *flagRegister8) get() uint8 {
	return reg.val
}

// register16 represents a CPU register of 16-bit size.
type register16 interface {
	// set sets the register's value.
	set(val uint16) uint16
	// get returns the register's value.
	get() uint16
}

// normalRegister16 represents a normal 16-bit register. Registers SP and PC
// are both of this type.
type normalRegister16 struct {
	val uint16
}

func (reg *normalRegister16) set(val uint16) uint16 {
	reg.val = val
	return reg.get()
}

func (reg *normalRegister16) get() uint16 {
	return reg.val
}

// registerCombined represents a 16-bit register backed by two 8-bit registers.
// Writes to this register are reflected in the two registers it is backed by.
type registerCombined struct {
	first, second register8
}

func (reg *registerCombined) set(val uint16) uint16 {
	reg.first.set(uint8(val >> 8))
	reg.second.set(uint8(val))

	return reg.get()
}

func (reg *registerCombined) get() uint16 {
	return (uint16(reg.first.get()) << 8) | uint16(reg.second.get())
}

func (reg *registerCombined) size() int {
	return 16
}

type registerType8Bit int

const (
	regA registerType8Bit = iota
	regB
	regC
	regD
	regE
	regH
	regL
	regF
)

type registerType16Bit int

const (
	regAF registerType16Bit = iota
	regBC
	regDE
	regHL

	regSP
	regPC
)
