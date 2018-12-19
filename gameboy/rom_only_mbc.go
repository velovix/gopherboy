package gameboy

import "fmt"

// romOnlyMBC is the basic memory bank controller. It can scarcely be called a
// memory bank controller at all since there's no switching. This MBC provides
// a single 16K ROM bank.
type romOnlyMBC struct {
	// romBank1 is the one and only ROM bank provided by this basic controller.
	romBank1 []uint8
	// ramBank0 is the built-in Game Boy RAM.
	ramBank0 []uint8
}

func newROMOnlyMBC(cartridgeData []uint8) *romOnlyMBC {
	m := &romOnlyMBC{
		romBank1: make([]uint8, 0x4000),
		ramBank0: make([]uint8, 0x2000),
	}

	// Load bank 1 cartridge data
	for i := bankedROMAddr; i < videoRAMAddr; i++ {
		m.romBank1[i-bankedROMAddr] = cartridgeData[i]
	}

	return m
}

// at provides access to the bank 1 ROM.
func (m *romOnlyMBC) at(addr uint16) uint8 {
	switch {
	case inRAMArea(addr):
		return m.ramBank0[addr-ramAddr]
	case inBankedROMArea(addr):
		return m.romBank1[addr-bankedROMAddr]
	case inBankedRAMArea(addr):
		if printInstructions {
			fmt.Printf("Warning: Read from banked RAM section at address %#x, "+
				"but the ROM-only MBC does not support banked RAM\n",
				addr)
		}
		return 0xFF
	default:
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a write to address %#x\n", addr))
	}
}

// set can update bank 0 RAM, but otherwise does not support any special
// operations like real MBCs do.
func (m *romOnlyMBC) set(addr uint16, val uint8) {
	switch {
	case inBank0ROMArea(addr) || inBankedROMArea(addr):
		if printInstructions {
			fmt.Printf("Warning: Ignoring write to ROM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	case inRAMArea(addr):
		m.ramBank0[addr-ramAddr] = val
	case inBankedRAMArea(addr):
		if printInstructions {
			fmt.Printf("Warning: Ignoring write to banked RAM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	default:
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a write to address %#x\n", addr))
	}
}
