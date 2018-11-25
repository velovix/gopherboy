package main

import "fmt"

// mbc3 implements the MBC3 memory bank controller. An MBC3 can support up to
// 128 16K ROM banks, up to up to 8 8K RAM banks, potentially a real time
// clock (RTC), and potentially a battery backup.
type mbc3 struct {
	// romBanks contains cartridge ROM banks from 1 to X, where X is the number
	// of ROM banks in the cartridge.
	romBanks map[int][]uint8
	// ramBanks is a map of all RAM banks, where the key is the RAM bank
	// number. These extra RAM banks are supplied by the cartridge.
	ramBanks map[int][]uint8

	// True if this MBC has a real time clock in it.
	hasRTC bool

	// The currently selected ROM bank.
	currROMBank uint8
	// The currently selected RAM bank.
	currRAMBank uint8
	// True if RAM turned on.
	ramAndRTCEnabled bool
	// True if the RTC is "latched", meaning that its value should be frozen in
	// memory.
	rtcLatched bool
}

func newMBC3(header romHeader, cartridgeData []uint8, hasRTC bool) *mbc3 {
	var m mbc3

	if hasRTC {
		panic("MBC3 devices with an RTC are not supported")
	}

	m.hasRTC = hasRTC

	m.romBanks = makeROMBanks(header.romSizeType, cartridgeData)
	m.ramBanks = makeRAMBanks(header.ramSizeType)

	// TODO(velovix): Should 1 be the default ROM bank?
	m.currROMBank = 1
	// TODO(velovix): Should 0 be the default RAM bank?
	m.currRAMBank = 0

	return &m
}

// at provides access to the MBC3 banked ROM, banked RAM, and real time clock.
func (m *mbc3) at(addr uint16) uint8 {
	if addr >= bankedROMAddr && addr < videoRAMAddr {
		// Banked ROM area
		return m.romBanks[int(m.currROMBank)][addr-bankedROMAddr]
	} else if addr >= bankedRAMAddr && addr < ramAddr {
		// Banked RAM or Real Time Clock register area
		if _, ok := m.ramBanks[int(m.currRAMBank)]; !ok {
			if printInstructions {
				fmt.Printf("Warning: Invalid read from nonexistent "+
					"RAM bank %v at address %#x\n", m.currRAMBank, addr)
			}
			return 0xFF
		} else {
			return m.ramBanks[int(m.currRAMBank)][addr-bankedRAMAddr]
		}
	} else {
		panic(fmt.Sprintf("MBC3 is unable to handle reads to address %#x", addr))
	}
}

func (m *mbc3) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		// The RAM and RTC enable/disable area. Used to turn on and off access
		// to banked RAM and the RTC. These two devices are positioned in the
		// same memory location, so another option (below) chooses which one is
		// actually exposed.
		lower, _ := split(val)
		// 0x0A is the magic number to turn these devices on
		m.ramAndRTCEnabled = lower == 0x0A
	} else if addr < 0x4000 {
		// ROM Bank Number "register"
		// This area is used to specify all 7 bits of the desired ROM bank
		// number, which the MBC will switch to.
		// This "register" is only 7 bits in size, get those 7 bits
		bank := val & 0x7F
		if bank == 0x0 {
			// A special case where ROM bank 0 is interpreted as bank 1, since
			// bank 0 is always available
			bank = 1
		}

		m.currROMBank = bank
	} else if addr < 0x6000 {
		// RAM Bank Number or RTC Register Select
		// Writing a value of 0x00 to 0x07 in this register will switch to the
		// RAM bank of that number. Writing a value of 0x08 to 0x0C will map
		// that corresponding RTC register to the RAM bank address space.
		if val <= 0x07 {
			m.currRAMBank = val
		} else if val <= 0x0C {
			panic("Attempt to control the RTC register, but RTC is not supported")
		} else {
			panic(fmt.Sprintf("Unexpected value in RAM Bank/RTC Register Select %#x", val))
		}
	} else if addr < 0x8000 {
		// Latch Clock Data Register
		// Writing a 0x01 to this area will lock the RTC at its current value
		// until a 0x00 is written. Note that the RTC itself continues to tick
		// even while this latch is set. Only the in-memory value remains
		// unchanged.
		m.rtcLatched = val == 0x01
		panic("Attempt to latch the RTC register, but RTC is not supported")
	}
}
