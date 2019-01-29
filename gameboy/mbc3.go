package gameboy

import "fmt"

// mbc3 implements the MBC3 memory bank controller. An MBC3 can support up to
// 128 16K ROM banks, up to up to 8 8K RAM banks, potentially a real time
// clock (RTC), and potentially a battery backup.
type mbc3 struct {
	// romBanks contains cartridge ROM banks, indexed by their bank number.
	romBanks [][]uint8
	// ramBanks contains all extra RAM banks, indexed by their bank number.
	// These extra RAM banks are supplied by the cartridge.
	ramBanks [][]uint8

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

func newMBC3(
	header romHeader,
	cartridgeData []uint8,
	hasRTC bool) *mbc3 {

	var m mbc3

	if hasRTC {
		panic("MBC3 devices with an RTC are not supported")
	}

	m.hasRTC = hasRTC

	m.romBanks = makeROMBanks(header.romSizeType, cartridgeData)
	m.ramBanks = makeRAMBanks(header.ramSizeType)

	m.currROMBank = 1
	m.currRAMBank = 0

	return &m
}

// at provides access to the MBC3 banked ROM, banked RAM, and real time clock.
func (m *mbc3) at(addr uint16) uint8 {
	switch {
	case inBank0ROMArea(addr):
		return m.romBanks[0][addr]
	case inBankedROMArea(addr):
		bank := m.currROMBank
		if bank == 0 {
			// Bank 0 is not directly selectable, map to bank 1 instead
			bank = 1
		}
		// If an out-of-bounds ROM bank is selected, the value will "wrap
		// around"
		bank %= uint8(len(m.romBanks))
		return m.romBanks[bank][addr-bankedROMAddr]
	case inBankedRAMArea(addr):
		// Banked RAM or Real Time Clock register area
		if m.ramAndRTCEnabled && len(m.ramBanks) > 0 {
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
		panic(fmt.Sprintf("MBC3 is unable to handle reads to address %#x", addr))
	}
}

// set can do many things with the MBC3.
//
// If the target address is within ROM, it will control some aspect of the MBC3
// (like switching banks or controlling the RTC), depending on the address
// itself and the given value.
//
// If the target address is within the RAM bank area, the selected RAM bank
// will be written to.
func (m *mbc3) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		// The RAM and RTC enable/disable area. Used to turn on and off access
		// to banked RAM and the RTC. These two devices are positioned in the
		// same memory location, so another option (below) chooses which one is
		// actually exposed.
		lower, _ := split(val)
		// 0x0A is the magic number to turn these devices on
		m.ramAndRTCEnabled = lower == 0x0A
		fmt.Printf("Write to RAM enable/disable %#x: %#x, %v\n", addr, val, m.ramAndRTCEnabled)
	} else if addr < 0x4000 {
		// ROM Bank Number "register"
		// This area is used to specify all 7 bits of the desired ROM bank
		// number, which the MBC will switch to.

		// This "register" is only 7 bits in size, get those 7 bits
		bank := val & 0x7F
		// This register cannot have 0x0 written to it. A write of 0x0 will be
		// interpreted as 0x1. This means that bank 0x0 is not inaccessible
		if bank == 0x00 {
			bank = 0x01
		}
		m.currROMBank = bank
	} else if addr < 0x6000 {
		// RAM Bank Number or RTC Register Select
		// Writing a value of 0x00 to 0x07 in this register will switch to the
		// RAM bank of that number. Writing a value of 0x08 to 0x0C will map
		// that corresponding RTC register to the RAM bank address space.
		if val <= 0x07 {
			m.currRAMBank = val
			fmt.Println("Switched to RAM bank", m.currRAMBank)
		} else if val <= 0x0C && m.hasRTC {
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
		if m.hasRTC {
			m.rtcLatched = val == 0x01
			fmt.Println("Warning: Attempt to latch the RTC register, but RTC is not supported")
		}
	} else if inBankedRAMArea(addr) {
		if m.ramAndRTCEnabled {
			m.ramBanks[m.currRAMBank][addr-bankedRAMAddr] = val
		} else {
			fmt.Printf("Attempt to write to banked RAM when RAM is disabled: At address %#x\n", addr)
		}
	} else {
		panic(fmt.Sprintf("The MBC3 should not have been notified of a write "+
			"to address %#x", addr))
	}
}

func (m *mbc3) dumpBatteryBackedRAM() []uint8 {
	var dump []uint8

	for _, bank := range m.ramBanks {
		for _, val := range bank {
			dump = append(dump, val)
		}
	}

	return dump
}

func (m *mbc3) loadBatteryBackedRAM(dump []uint8) {
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
