package main

import "fmt"

// mbc1 implements an MBC1 memory bank controller. An MBC1 can support up to
// 125 16K ROM banks, up to 4 8K RAM banks, and potentially a battery backup.
type mbc1 struct {
	// romBanks contains cartridge ROM banks from 1 to X, where X is the number
	// of ROM banks in the cartridge.
	romBanks map[uint8][]uint8
	// ramBanks is a map of all RAM banks, where the key is the RAM bank
	// number. These extra RAM banks are supplied by the cartridge.
	ramBanks map[uint8][]uint8

	// The currently selected ROM bank.
	currROMBank uint8
	// The currently selected RAM bank.
	currRAMBank uint8
	// True if RAM turned on.
	ramEnabled bool
	// Controls how any bits written to 0x4000-0x6000 are interpreted. If this
	// value is 0, they are interpreted as the upper bits of the ROM bank
	// selection, where the lower bits are whatever is written to
	// 0x4000-0x8000. If this value is 1, they are interpreted as the RAM bank
	// selection.
	bankSelectionMode uint8
}

func newMBC1(header romHeader, cartridgeData []uint8) *mbc1 {
	var m mbc1

	// Create ROM banks
	switch header.romSizeType {
	case 0x00:
		// A single ROM bank, no switching going on
		m.romBanks = map[uint8][]uint8{
			1: make([]uint8, 0x4000),
		}
	case 0x01:
		// Three switchable ROM banks
		m.romBanks = map[uint8][]uint8{
			1: make([]uint8, 0x4000),
			2: make([]uint8, 0x4000),
			3: make([]uint8, 0x4000),
		}
	default:
		panic(fmt.Sprintf("Unsupported ROM size type %v", header.romSizeType))
	}
	// Put cartridge data into each bank
	for bank, data := range m.romBanks {
		startAddr := 0x4000 * int(bank)
		for i := startAddr; i < startAddr+0x4000; i++ {
			data[i-startAddr] = cartridgeData[i]
		}
	}

	// TODO(velovix): Should 1 be the default ROM bank?
	m.currROMBank = 1

	// Create RAM banks
	switch header.ramSizeType {
	case 0x00:
		// No bank
		m.ramBanks = map[uint8][]uint8{}
	case 0x01:
		// One 2 KB bank
		m.ramBanks = map[uint8][]uint8{
			1: make([]uint8, 2000),
		}
	case 0x02:
		// One 8 KB bank
		m.ramBanks = map[uint8][]uint8{
			1: make([]uint8, 8000),
		}
	case 0x03:
		// Four 8 KB banks
		m.ramBanks = make(map[uint8][]uint8)
		for i := uint8(1); i <= 4; i++ {
			m.ramBanks[i] = make([]uint8, 8000)
		}
	case 0x04:
		// Sixteen 8 KB banks
		m.ramBanks = make(map[uint8][]uint8)
		for i := uint8(1); i <= 16; i++ {
			m.ramBanks[i] = make([]uint8, 8000)
		}
	case 0x05:
		// Eight 8 KB banks
		m.ramBanks = make(map[uint8][]uint8)
		for i := uint8(1); i <= 8; i++ {
			m.ramBanks[i] = make([]uint8, 8000)
		}
	default:
		panic(fmt.Sprintf("Unknown RAM size type %v", header.ramSizeType))
	}

	// TODO(velovix): Should 0 be the default RAM bank?
	m.currRAMBank = 0

	return &m
}

// at provides access to the MBC1 banked ROM and RAM.
func (m *mbc1) at(addr uint16) uint8 {
	if addr >= bankedROMAddr && addr < videoRAMAddr {
		// Banked ROM area
		return m.romBanks[m.currROMBank][addr-bankedROMAddr]
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		// Banked RAM area
		return m.ramBanks[m.currRAMBank][addr-bankedRAMAddr]
	} else {
		panic(fmt.Sprintf("It isn't the MBC1's job to handle access to address %#x", addr))
	}
}

// set can do many things with the MBC1.
//
// If the target address is within ROM, it will control some aspect of the MBC1
// (like switching banks), depending on the address itself and the given value.
//
// If the target address is within the RAM bank area, the selected RAM bank
// will be written to.
func (m *mbc1) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		// The RAM enable/disable area. Used to turn on and off access to
		// banked RAM.
		lower, _ := split(val)
		m.ramEnabled = lower == 0x0A
	} else if addr < 0x4000 {
		// ROM Bank 01-7F "register"
		// The area is used to specify the lower 5 bits of the desired ROM bank
		// number.
		// This "register" is only 5 bits in size, get those 5 bits
		bank := val & 0x1F
		if bank == 0x0 {
			// A special case where ROM bank 0 is interpreted as bank 1, since
			// bank 0 is always available
			bank = 1
		} else if bank == 0x20 || bank == 0x40 || bank == 0x60 {
			// A special case where ROM banks 0x20, 0x40, and 0x60 don't exist
			// and instead map to the bank directly after them
			bank++
		}

		// Replace the first 5 bits of the current ROM bank with these
		m.currROMBank &= 0xE0
		m.currROMBank |= bank
		fmt.Println("Switching to ROM bank", m.currROMBank)
	} else if addr < 0x6000 {
		// RAM Bank Number or Upper Bits of ROM Bank Number "register"
		// This area is used to specify the upper 2 bits of the desired ROM
		// bank number OR the only 2 bits used to specify the desired RAM bank
		// number. The desired behavior can be selected using the ROM/RAM Mode
		// Select "register", specified below.
		// This "register" is only 2 bits in size, get those 2 bits
		bank := val & 0x03
		if m.bankSelectionMode == 1 {
			fmt.Println("Switching to RAM bank", bank)
			panic("... but this isn't supported")
		} else {
			// Set bits 5 and 6 of the current ROM bank
			m.currROMBank &= 0x9F
			m.currROMBank |= bank << 5
		}
	} else if addr < 0x8000 {
		// The ROM/RAM Mode Select "register"
		// This changes which mode the RAM Bank Number or Upper Bits of ROM
		// Bank Number "register" uses.
		m.bankSelectionMode = val & 0x01
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		// Banked RAM
		if m.ramEnabled {
			m.ramBanks[m.currRAMBank][addr-bankedRAMAddr] = val
		} else {
			panic("Attempt to write to banked RAM when RAM is disabled")
		}
	} else {
		panic(fmt.Sprintf("It isn't the MBC1's job to handle writes to address %#x", addr))
	}
}
