package main

type mbc interface {
	// at returns the value found at the given address.
	at(addr uint16) uint8
	// atRange returns a slice of values found in the range [start:end).
	atRange(start, end uint16) []uint8
	// set sets the value at the given address to the given value.
	set(addr uint16, val uint8)
	// pointerTo returns a pointer to the place in memory specified by the
	// given address. Edits can be made to this pointer's value later and it
	// will be reflected in memory.
	pointerTo(addr uint16) *uint8
	// dump returns a dump of the current memory.
	dump() []byte
}

// romOnlyMBC isn't actually a memory bank controller. It's for games that only
// use bank 0 of ROM with no actual MBC of their own.
type romOnlyMBC struct {
	mem []uint8
}

func newROMOnlyMBC(cartridgeData []byte) romOnlyMBC {
	m := romOnlyMBC{
		mem: make([]uint8, 0x10000),
	}

	// Put cartridge data into bank 0
	for i := 0; i < romBank0End; i++ {
		m.mem[i] = cartridgeData[i]
	}

	return m
}

func (m romOnlyMBC) at(addr uint16) uint8 {
	return m.mem[addr]
}

func (m romOnlyMBC) atRange(start, end uint16) []uint8 {
	return m.mem[start:end]
}

func (m romOnlyMBC) set(addr uint16, val uint8) {
	if addr <= romBank0End {
		panic("Attempt to write to ROM space")
	}
	m.mem[addr] = val
}

func (m romOnlyMBC) pointerTo(addr uint16) *uint8 {
	return &m.mem[addr]
}

func (m romOnlyMBC) dump() []byte {
	output := make([]byte, len(m.mem))
	copy(output, m.mem)
	return output
}

// mbc1 represents a memory bank controller that supports ROM bank switching
// and external battery-backed RAM for game saves, if available.
type mbc1 struct {
	// romBank0 refers to the first 16 KB of ROM, which is always assigned to
	// bank 0.
	romBank0 []uint8
	// romBankX holds memory from the bank that was last switched to.
	romBankX []uint8
	// systemMem refers to the area after 0x8000 that is designated for non-ROM
	// purposes.
	// TODO(velovix): Come up with a better name for this.
	systemMem []uint8
	// The currently selected ROM bank.
	romBank int
	// The currently selected RAM bank.
	ramBank int
	// True if RAM turned on.
	ramEnabled bool
	// The full cartridge dump. Consulted for when bank switching occurs.
	cartridgeData []byte
}

const mbc1BankBytes = 0x4000

func newMBC1(cartridgeData []byte) *mbc1 {
	m := &mbc1{
		romBank0:      make([]uint8, romBank0End),
		systemMem:     make([]uint8, 0x10000-videoRAMStart),
		cartridgeData: cartridgeData,
	}

	// Put cartridge data into bank 0
	for i := 0; i < romBank0End; i++ {
		m.romBank0[i] = cartridgeData[i]
	}

	return m
}

func (m *mbc1) at(addr uint16) uint8 {
	if addr < romBank0End {
		return m.romBank0[addr]
	} else if addr >= romBank0End && addr < videoRAMStart {
		if m.romBankX == nil {
			// TODO(velovix): Is this actually illegal? If not, what is the
			// default bank?
			panic("attempt to access ROM bank area with no bank selected")
		} else {
			return m.romBankX[addr-romBank0End]
		}
	} else {
		return m.systemMem[addr-videoRAMStart]
	}
}

func (m *mbc1) atRange(start, end uint16) []uint8 {
	if start < romBank0End && end >= romBank0End ||
		start < videoRAMStart && end >= videoRAMStart {
		panic("crossing RAM boundaries in atRange is not yet supported")
	}
	if start < romBank0End {
		return m.romBank0[start:end]
	} else if start >= romBank0End && start < videoRAMStart {
		if m.romBankX == nil {
			// TODO(velovix): Is this actually illegal? If not, what is the
			// default bank?
			panic("attempt to access ROM bank area with no bank selected")
		} else {
			return m.romBankX[start-romBank0End : end-romBank0End]
		}
	} else {
		return m.systemMem[start-videoRAMStart : end-videoRAMStart]
	}
}

func (m *mbc1) set(addr uint16, val uint8) {
	if addr < 0x2000 {
		panic("Turning RAM on is not supported")
	} else if addr >= 0x2000 && addr < 0x4000 {
		// Switch banks
		if val == 0x0 {
			// A special case where ROM bank 0 is interpreted as bank 1, since
			// bank 0 is always available
			val = 1
		} else if val == 0x20 || val == 0x40 || val == 0x60 {
			// A special case where ROM banks 0x20, 0x40, and 0x60 don't exist
			// and instead map to the bank directly after them
			val++
		}

		bankStart := mbc1BankBytes * int(val)
		m.romBankX = m.cartridgeData[bankStart : bankStart+mbc1BankBytes]
	} else if addr >= 0x4000 && addr < 0x6000 {
		// TODO(velovix): Figure out in what order the upper and lower bits of
		// the ROM bank selection are supposed to be written
		panic("Selecting RAM banks and selecting ROM banks higher than 31 " +
			"is not supported")
	} else if addr >= 0x6000 && addr < 0x8000 {
		panic("RAM bank switching is not supported, so it can't be turned on")
	} else {
		// Set some in-memory register value or whatever
		m.systemMem[addr-videoRAMStart] = val
	}
}

func (m *mbc1) pointerTo(addr uint16) *uint8 {
	if addr < romBank0End {
		return &m.romBank0[addr]
	} else if addr >= romBank0End && addr < videoRAMStart {
		return &m.romBankX[addr-romBank0End]
	} else {
		return &m.systemMem[addr-videoRAMStart]
	}
}

func (m *mbc1) dump() []byte {
	return append(m.romBank0, append(m.romBankX, m.systemMem...)...)
}
