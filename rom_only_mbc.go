package main

import "fmt"

// romOnlyMBC is the basic memory bank controller. It can scarcely be called a
// memory bank controller at all since there's no switching. This MBC provides
// a single 16K ROM bank.
type romOnlyMBC struct {
	// romBank1 is the one and only ROM bank provided by this basic controller.
	romBank1 []uint8
}

func newROMOnlyMBC(cartridgeData []uint8) *romOnlyMBC {
	m := &romOnlyMBC{romBank1: make([]uint8, 0x4000)}

	// Load bank 1 cartridge data
	for i := bankedROMAddr; i < videoRAMAddr; i++ {
		m.romBank1[i-bankedROMAddr] = cartridgeData[i]
	}

	return m
}

// at provides access to the bank 1 ROM.
func (m *romOnlyMBC) at(addr uint16) uint8 {
	if addr >= bankedROMAddr && addr < videoRAMAddr {
		return m.romBank1[addr-bankedROMAddr]
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		fmt.Printf("Warning: Read from banked RAM section at address %#x, "+
			"but the ROM-only MBC does not support banked RAM\n",
			addr)
		// TODO(velovix): Is this the correct behavior?
		return 0xFF
	} else {
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a write to address %#x\n", addr))
	}
}

// set does nothing, since there are no write operations that this MBC
// supports.
func (m *romOnlyMBC) set(addr uint16, val uint8) {
	if addr < videoRAMAddr {
		if printInstructions {
			fmt.Printf("Warning: Ignoring write to ROM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		if printInstructions {
			fmt.Printf("Warning: Ignoring write to banked RAM space "+
				"at %#x with ROM-only MBC\n", addr)
		}
	} else {
		panic(fmt.Sprintf("The ROM-only MBC should not have been "+
			"notified of a write to address %#x\n", addr))
	}
}
