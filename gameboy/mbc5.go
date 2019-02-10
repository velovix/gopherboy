package gameboy

import "fmt"

type mbc5 struct {
	// romBanks contains cartridge ROM banks, indexed by their bank number.
	romBanks [][]uint8
	// ramBanks contains all extra RAM banks, indexed by their bank number.
	// These extra RAM banks are supplied by the cartridge.
	ramBanks [][]uint8

	// The currently selected ROM bank.
	currROMBank uint16
	// The currently selected RAM bank.
	currRAMBank uint8
	// True if RAM turned on.
	ramEnabled bool

	// If true, then this cartridge is equipped with rumble.
	hasRumble bool
	// If true, the rumble motor is currently running.
	rumbling bool
}

func newMBC5(
	header romHeader,
	cartridgeData []uint8,
	hasRumble bool) *mbc5 {

	var m mbc5

	m.romBanks = makeROMBanks(header.romSizeType, cartridgeData)
	m.ramBanks = makeRAMBanks(header.ramSizeType)

	m.currROMBank = 1
	m.currRAMBank = 0

	m.hasRumble = hasRumble

	return &m
}

// at provides access to the MBC5 banked ROM and RAM.
func (m *mbc5) at(addr uint16) uint8 {
	switch {
	case inBank0ROMArea(addr):
		return m.romBanks[0][addr]
	case inBankedROMArea(addr):
		bank := m.currROMBank
		// If an out-of-bounds ROM bank is selected, the value will "wrap
		// around"
		bank %= uint16(len(m.romBanks))
		return m.romBanks[bank][addr-bankedROMAddr]
	case inBankedRAMArea(addr):
		if m.ramEnabled && len(m.ramBanks) > 0 {
			bank := m.currRAMBank
			// If an out-of-bounds RAM bank is selected, the value will "wrap
			// around"
			bank %= uint8(len(m.ramBanks))

			return m.ramBanks[bank][addr-bankedRAMAddr]
		} else {
			// The default value for disabled or unavailable RAM
			return 0xFF
		}
	default:
		panic(fmt.Sprintf("MBC5 is unable to handle reads to address %#x", addr))
	}
}

// set can do many things with the MBC5.
//
// Writing to special areas in ROM can turn on and off cartridge RAM, specify
// the current ROM bank, specify the current RAM bank, and turn on and off
// rumble.
//
// If the target address is within the RAM bank area, the selected RAM bank
// will be written to.
func (m *mbc5) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		// The RAM enable/disable area. Used to turn on and off access to
		// banked RAM.
		lower, _ := split(val)
		// 0x0A is the magic number to turn this device on
		m.ramEnabled = lower == 0x0A
	} else if addr < 0x3000 {
		// ROM Bank (Low bits)
		// This are is used to specify the lower 8 bits of the desired ROM bank
		// number, which the MBC will switch to.
		_, upper := split16(m.currROMBank)
		m.currROMBank = combine16(val, upper)
	} else if addr < 0x4000 {
		// ROM Bank (High bits)
		// This are is used to specify the upper 8 bits of the desired ROM bank
		// number, which the MBC will switch to.
		// The Cycle-Accurate Game Boy Docs claim that this area is a mirror to
		// ROM Bank (Low bits) when fewer than 256 banks are used. I'm doubtful
		// but who am I to say.
		if len(m.romBanks) < 256 {
			// This area is just a mirror for ROM Bank (Low bits)
			_, upper := split16(m.currROMBank)
			m.currROMBank = combine16(val, upper)
		} else {
			// Actually set the high bits of the ROM bank
			lower, _ := split16(m.currROMBank)
			m.currROMBank = combine16(lower, val)
		}
	} else if addr < 0x6000 {
		// RAM Bank and Rumble Enable
		if m.hasRumble {
			// Rumble on/off is mapped to bit 4 on this register. As a result,
			// only bits 0-3 are mapped to RAM bank switching, limiting the
			// maximum number of RAM banks these cartridges can have.
			m.currRAMBank = val & 0x0F
			m.rumbling = val&0x10 == 0x10
		} else {
			m.currRAMBank = val
		}
	} else if addr < 0x8000 {
		// This area of ROM doesn't do anything when written to
	} else if inBankedRAMArea(addr) {
		if m.ramEnabled {
			m.ramBanks[m.currRAMBank][addr-bankedRAMAddr] = val
		} else {
			fmt.Printf("Attempt to write to banked RAM when RAM is disabled: At address %#x\n", addr)
		}
	} else {
		panic(fmt.Sprintf("MBC5 is unable to handle writes to address %#x", addr))
	}
}

func (m *mbc5) dumpBatteryBackedRAM() []uint8 {
	var dump []uint8

	for _, bank := range m.ramBanks {
		for _, val := range bank {
			dump = append(dump, val)
		}
	}

	return dump
}

func (m *mbc5) loadBatteryBackedRAM(dump []uint8) {
	for bankNum, bank := range m.ramBanks {
		for i := range bank {
			dumpIndex := len(bank)*bankNum + i
			if i >= len(dump) {
				panic(fmt.Sprintf("RAM dump is too small for this MBC: %v", i))
			}
			bank[i] = dump[dumpIndex]
		}
	}
}
