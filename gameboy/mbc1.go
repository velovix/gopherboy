package gameboy

import "fmt"

// mbc1 implements an MBC1 memory bank controller. An MBC1 can support up to
// 125 16K ROM banks, up to 4 8K RAM banks, and potentially a battery backup.
type mbc1 struct {
	// romBanks contains cartridge ROM banks from 1 to X, where X is the number
	// of ROM banks in the cartridge.
	romBanks map[int][]uint8
	// ramBanks is a map of all RAM banks, where the key is the RAM bank
	// number. These extra RAM banks are supplied by the cartridge.
	ramBanks map[int][]uint8

	// bankReg1 is set by writing to 0x2000-0x3FFF. It is a 5-bit value. It is
	// used to specify the lower 5 bits of the desired ROM bank.
	bankReg1 uint8
	// bankReg2 is set by writing to 0x6000-0x7FFF. It is a 2-bit value. In
	// bank selection mode 0, it is used to specify the upper 2 bits of the
	// desired ROM bank in 0x4000-0x7FFF. In bank selection mode 1, it is used
	// to specify the upper 2 bits of the desired ROM bank in 0x0000-0x3FFF and
	// the lower 2 bits of the desired RAM bank.
	bankReg2 uint8

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

	m.romBanks = makeROMBanks(header.romSizeType, cartridgeData)
	m.ramBanks = makeRAMBanks(header.ramSizeType)

	// The default bank values
	m.bankReg1 = 0x1
	m.bankReg2 = 0x0

	return &m
}

// at provides access to the MBC1 banked ROM and RAM.
func (m *mbc1) at(addr uint16) uint8 {
	switch {
	case inBank0ROMArea(addr):
		switch m.bankSelectionMode {
		case 0:
			// This area is always bank 0 in this mode
			return m.romBanks[0][addr]
		case 1:
			// The bank here is specified by the value written to
			// 0x4000-0x5FFF, shifted 5 bits over.
			bank := int(m.bankReg2 << 5)
			// If an out-of-bounds ROM bank is selected, the value will "wrap
			// around"
			bank %= len(m.romBanks)
			return m.romBanks[bank][addr]
		default:
			panic(fmt.Sprintf("invalid bank selection mode %#x", m.bankSelectionMode))
		}
	case inBankedROMArea(addr):
		// The current bank is calculated by combining bank registers 1 and 2
		bank := int(m.bankReg1 | (m.bankReg2 << 5))
		// If an out-of-bounds ROM bank is selected, the value will "wrap
		// around"
		bank %= len(m.romBanks)

		return m.romBanks[bank][addr-bankedROMAddr]
	case inBankedRAMArea(addr):
		if m.ramEnabled && len(m.ramBanks) > 0 {
			switch m.bankSelectionMode {
			case 0:
				// In this mode, bank 0 is always used
				return m.ramBanks[0][addr-bankedRAMAddr]
			case 1:
				// The current bank is equal to the value of bank register 2
				bank := int(m.bankReg2)
				// If an out-of-bounds RAM bank is selected, the value will "wrap
				// around"
				bank %= len(m.ramBanks)

				return m.ramBanks[bank][addr-bankedRAMAddr]
			default:
				panic(fmt.Sprintf("invalid bank selection mode %v", m.bankSelectionMode))
			}
		} else {
			// The default value for disabled RAM
			return 0xFF
		}
	default:
		panic(fmt.Sprintf("MBC1 is unable to handle reads to address %#x", addr))
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
		// 0x0A is the magic number to turn this device on
		m.ramEnabled = lower == 0x0A
	} else if addr < 0x4000 {
		// ROM Bank 01-7F register
		// This area controls the value of what I call "bank register 1". For
		// more information on what this does, see the field's documentation on
		// the mbc1 type.
		m.bankReg1 = val & 0x1F

		// This register cannot have 0x0 written to it. A write of 0x0 will be
		// interpreted as 0x1. This means that banks 0x0, 0x20, 0x40, and 0x60
		// are inaccessible
		if m.bankReg1 == 0x00 {
			m.bankReg1 = 0x01
		}
	} else if addr < 0x6000 {
		// RAM Bank Number or Upper Bits of ROM Bank Number "register"
		// This area controls the value of what I call "bank register 2". For
		// more information on what this does, see the field's documentation on
		// the mbc1 type.
		// This "register" is only 2 bits in size, get those 2 bits
		m.bankReg2 = val & 0x03
	} else if addr < 0x8000 {
		// The ROM/RAM Mode Select "register"
		// This changes which mode the RAM Bank Number or Upper Bits of ROM
		// Bank Number "register" uses.
		m.bankSelectionMode = val & 0x01
	} else if addr >= ramAddr && addr < ramMirrorAddr {
		// Bank 0 RAM area
		m.ramBanks[0][addr-ramAddr] = val
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		// Banked RAM
		if len(m.ramBanks) > 0 && m.ramEnabled {
			switch m.bankSelectionMode {
			case 0:
				// In this mode, bank 0 is always used
				m.ramBanks[0][addr-bankedRAMAddr] = val
			case 1:
				// The current RAM bank is controlled by bank register 2
				bank := int(m.bankReg2)
				// If an out-of-bounds RAM bank is selected, the value will "wrap
				// around"
				bank %= len(m.ramBanks)
				m.ramBanks[bank][addr-bankedRAMAddr] = val
			default:
				panic(fmt.Sprintf("invalid bank selection mode %v", m.bankSelectionMode))
			}
		}
	} else {
		panic(fmt.Sprintf("It isn't the MBC1's job to handle writes to address %#x", addr))
	}
}
