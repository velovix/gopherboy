package main

import "fmt"

// mmu describes the interface a memory management unit provides. A memory
// management unit provides access to the memory. Reading or writing to memory
// may have special effects depending on the memory bank controller being used.
type mmu interface {
	// at returns the value found at the given address.
	at(addr uint16) uint8
	// set sets the value at the given address to the given value.
	set(addr uint16, val uint8)
	// dump returns a dump of the current memory.
	dump() []byte
}

// romOnly is a simple memory management unit for games that only use bank 0 of
// ROM with no additional ROMs or special devices.
type romOnly struct {
	mem []uint8
}

func newROMOnly(cartridgeData []byte) romOnly {
	m := romOnly{
		mem: make([]uint8, 0x10000),
	}

	// Put cartridge data into bank 0
	for i := 0; i < bankedROMAddr; i++ {
		m.mem[i] = cartridgeData[i]
	}

	return m
}

func (m romOnly) at(addr uint16) uint8 {
	return m.mem[addr]
}

func (m romOnly) set(addr uint16, val uint8) {
	if addr < videoRAMAddr {
		panic(fmt.Sprintf("Attempt to write to ROM space: %#x", addr))
	}
	m.mem[addr] = val
}

func (m romOnly) dump() []byte {
	output := make([]byte, len(m.mem))
	copy(output, m.mem)
	return output
}

// mbc1 represents a memory bank controller that supports ROM bank switching
// and external battery-backed RAM for game saves, if available.
type mbc1 struct {
	// bank0ROM refers to the first 16 KB of ROM, which is always assigned to
	// bank 0.
	bank0ROM []uint8
	// romBanks contains cartridge ROM banks from 1 to X, where X is the number
	// of ROM banks in the cartridge.
	romBanks map[uint8][]uint8
	// videoRAM is where sprite and tile data is stored for the video
	// controller to access.
	videoRAM []uint8
	// ramBanks is a map of all RAM banks, where the key is the RAM bank
	// number. These extra RAM banks are supplied by the cartridge.
	ramBanks map[uint8][]uint8
	// ram refers to the place in memory where the Game Boy's internal RAM is
	// stored.
	ram []uint8
	// oamRAM is where sprite attribute data is stored for the video controller
	// to access.
	oamRAM []uint8
	// ioRAM is where memory registers are mapped to for controlling and
	// reading various aspects of the hardware.
	ioRAM []uint8
	// hram is a special general-purpose RAM area
	hram []uint8

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

func newMBC1(cartridgeData []byte, header romHeader) *mbc1 {
	var m mbc1

	// Put cartridge data into bank 0
	m.bank0ROM = make([]uint8, bankedROMAddr-bank0ROMAddr)
	for i := 0; i < bankedROMAddr; i++ {
		m.bank0ROM[i] = cartridgeData[i]
	}

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

	m.videoRAM = make([]uint8, bankedRAMAddr-videoRAMAddr)

	// Create RAM banks
	m.ram = make([]uint8, 0x2000)
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

	m.ram = make([]uint8, invalidArea1Addr-ramAddr)
	m.oamRAM = make([]uint8, invalidArea2Addr-oamRAMAddr)
	m.ioRAM = make([]uint8, hramAddr-ioAddr)
	m.hram = make([]uint8, 0x10000-hramAddr)

	return &m
}

func (m *mbc1) at(addr uint16) uint8 {
	if addr < bankedROMAddr {
		// Bank 0 ROM
		return m.bank0ROM[addr-bank0ROMAddr]
	} else if addr < videoRAMAddr {
		// Banked ROM
		return m.romBanks[m.currROMBank][addr-bankedROMAddr]
	} else if addr < bankedRAMAddr {
		// Video RAM
		return m.videoRAM[addr-videoRAMAddr]
	} else if addr < ramAddr {
		// Banked RAM
		return m.ramBanks[m.currRAMBank][addr-bankedRAMAddr]
	} else if addr < invalidArea1Addr {
		// General-purpose bank 0 RAM address
		return m.ram[addr-ramAddr]
	} else if addr < oamRAMAddr {
		// Invalid area
		// TODO(velovix): Should any special behavior be exhibited here?
		return 0
	} else if addr < invalidArea2Addr {
		// OAM RAM address
		return m.oamRAM[addr-oamRAMAddr]
	} else if addr < ioAddr {
		// Invalid area
		// TODO(velovix): Should any special behavior be exhibited here?
		return 0
	} else if addr < hramAddr {
		// IO address
		return m.ioRAM[addr-ioAddr]
	} else if addr <= 0xFFFF {
		// HRAM address
		return m.hram[addr-hramAddr]
	} else {
		panic(fmt.Sprintf("Unexpected memory address %#x", addr))
	}
}

func (m *mbc1) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		lower, _ := split(val)
		if lower == 0x0A {
			m.ramEnabled = true
		} else {
			m.ramEnabled = false
		}
	} else if addr >= 0x2000 && addr < 0x4000 {
		// Switch banks
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
	} else if addr >= 0x4000 && addr < 0x6000 {
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
	} else if addr >= 0x6000 && addr < 0x8000 {
		// Switch bank selection modes
		m.bankSelectionMode = val & 0x01
	} else if addr < bankedRAMAddr {
		// Video RAM
		m.videoRAM[addr-videoRAMAddr] = val
	} else if addr < ramAddr {
		// Banked RAM
		if m.ramEnabled {
			m.ramBanks[m.currRAMBank][addr-bankedRAMAddr] = val
		} else {
			panic("Attempt to write to banked RAM when RAM is disabled")
		}
	} else if addr < invalidArea1Addr {
		// General-purpose RAM
		m.ram[addr-ramAddr] = val
	} else if addr < oamRAMAddr {
		// Invalid address
		// TODO(velovix): Should any special behavior be exhibited here?
	} else if addr < invalidArea2Addr {
		// OAM RAM
		m.oamRAM[addr-oamRAMAddr] = val
	} else if addr < ioAddr {
		// Invalid address
		// TODO(velovix): Should any special behavior be exhibited here?
	} else if addr < hramAddr {
		// I/O RAM
		m.ioRAM[addr-ioAddr] = val
	} else if addr <= 0xFFFF {
		// HRAM
		m.hram[addr-hramAddr] = val
	} else {
		panic(fmt.Sprintf("Unexpected memory address %#x", addr))
	}
}

func (m *mbc1) dump() []byte {
	// TODO(velovix): Implement this
	return make([]uint8, 0)
}

const (
	bank0ROMAddr     = 0x0000
	bankedROMAddr    = 0x4000
	videoRAMAddr     = 0x8000
	bankedRAMAddr    = 0xA000
	ramAddr          = 0xC000
	invalidArea1Addr = 0xE000
	oamRAMAddr       = 0xFE00
	invalidArea2Addr = 0xFEA0
	ioAddr           = 0xFF00
	hramAddr         = 0xFF80
)
