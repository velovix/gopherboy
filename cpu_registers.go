package main

// register represents a CPU register of either 8-bit or 16-bit size.
type register interface {
	// set sets the register's value. If the register is too small for the
	// value, it will be truncated bit-wise. Returns the set result.
	set(val uint16) uint16
	// get returns the register's value.
	get() uint16
	// size returns the size in bits of the register.
	size() int
}

// register8Bit represents a normal 8-bit register. Registers A, B, C, D, E, H,
// L, and F are all of this type.
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

// register16Bit represents a normal 16-bit register. Registers SP and PC are
// both of this type.
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

// registerCombined represents a 16-bit register backed by two 8-bit registers.
// Writes to this register are reflected in the two registers it is backed by.
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
