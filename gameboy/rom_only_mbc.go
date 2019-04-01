package gameboy

import "fmt"

// romOnlyMBC is the basic memory bank controller. It can scarcely be called a
// memory bank controller at all since there's no switching. This MBC provides
// a single 16K ROM bank.
type romOnlyMBC struct {
	// romBanks are the two available ROM banks in this basic controller.
	romBanks [][]uint8
}

func newROMOnlyMBC(header romHeader, cartridgeData []uint8) *romOnlyMBC {
	m := &romOnlyMBC{
		romBanks: makeROMBanks(header.romSizeType, cartridgeData),
	}

	return m
}

// at provides access to the bank 1 ROM.
func (m *romOnlyMBC) at(addr uint16) uint8 {
	switch {
	case inBank0ROMArea(addr):
		return m.romBanks[0][addr-bank0ROMAddr]
	case inBankedROMArea(addr):
		return m.romBanks[1][addr-bankedROMAddr]
	case inBankedRAMArea(addr):
		if printWarnings {
			fmt.Printf("Warning: Read from banked RAM section at address %#x, "+
				"but the ROM-only MBC does not support banked RAM\n",
				addr)
		}
		return 0xFF
	default:
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a read to address %#x\n", addr))
	}
}

// set can update bank 0 RAM, but otherwise does not support any special
// operations like real MBCs do.
func (m *romOnlyMBC) set(addr uint16, val uint8) {
	switch {
	case inBank0ROMArea(addr) || inBankedROMArea(addr):
		if printWarnings {
			fmt.Printf("Warning: Ignoring write to ROM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	case inBankedRAMArea(addr):
		if printWarnings {
			fmt.Printf("Warning: Ignoring write to banked RAM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	default:
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a write to address %#x\n", addr))
	}
}
