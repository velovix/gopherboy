package main

import (
	"fmt"
)

// dmaCycleLength is the number of cycles a DMA transfer takes.
const dmaCycleLength = 671

// mmu is the memory management unit. It handles all operations that are common
// to all Game Boy games and defers to the cartridge's memory bank controller
// in cartridge-specific cases.
type mmu struct {
	// bank0ROM refers to the first 16 KB of ROM, which is always assigned to
	// bank 0.
	bank0ROM []uint8
	// videoRAM is where sprite and tile data is stored for the video
	// controller to access.
	videoRAM []uint8
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

	// dmaActive is true if a DMA transfer is happening.
	dmaActive bool
	// dmaCursor is the next memory address to be transferred.
	dmaCursor uint16
	// dmaCycleCount is the number of a cycles a DMA transfer takes.
	dmaCycleCount int

	// subscribers maps a memory address to a function that should be called
	// when that memory address is written to. The value written to memory may
	// be modified by this function.
	subscribers map[uint16]onWriteFunc

	// mbc is the memory bank controller that this MMU will use.
	mbc mbc

	db *debugger
}

type onWriteFunc func(addr uint16, val uint8) uint8

func newMMU(cartridgeData []uint8, mbc mbc) *mmu {
	m := &mmu{
		videoRAM:    make([]uint8, bankedRAMAddr-videoRAMAddr),
		ram:         make([]uint8, ramMirrorAddr-ramAddr),
		oamRAM:      make([]uint8, invalidArea2Addr-oamRAMAddr),
		ioRAM:       make([]uint8, hramAddr-ioAddr),
		hram:        make([]uint8, lastAddr-hramAddr+1),
		subscribers: make(map[uint16]onWriteFunc),
		mbc:         mbc,
	}

	// Put cartridge data into bank 0
	m.bank0ROM = make([]uint8, bankedROMAddr-bank0ROMAddr)
	for i := 0; i < bankedROMAddr; i++ {
		m.bank0ROM[i] = cartridgeData[i]
	}

	m.subscribeTo(dmaAddr, m.onDMAWrite)

	return m
}

func (m *mmu) at(addr uint16) uint8 {
	if m.db != nil {
		m.db.memReadHook(addr)
	}

	if addr < bankedROMAddr {
		// Bank 0 ROM
		return m.bank0ROM[addr-bank0ROMAddr]
	} else if addr < videoRAMAddr {
		// Some additional ROM bank, controlled by the MBC
		return m.mbc.at(addr)
	} else if addr < bankedRAMAddr {
		// Video RAM
		return m.videoRAM[addr-videoRAMAddr]
	} else if addr < ramAddr {
		// Some additional RAM bank, controlled by the MBC
		return m.mbc.at(addr)
	} else if addr < ramMirrorAddr {
		// General-purpose bank 0 RAM
		return m.ram[addr-ramAddr]
	} else if addr < oamRAMAddr {
		// A bank 0 RAM mirror
		return m.ram[addr-ramMirrorAddr]
	} else if addr < invalidArea2Addr {
		// OAM RAM address
		return m.oamRAM[addr-oamRAMAddr]
	} else if addr < ioAddr {
		// Invalid area, which always returns 0xFF since it's the MMU's default
		// value
		if printInstructions {
			fmt.Printf("Warning: Read from invalid memory address %#x\n", addr)
		}
		return 0xFF
	} else if addr < hramAddr {
		// IO address space
		return m.ioRAM[addr-ioAddr]
	} else if addr <= lastAddr {
		// HRAM address
		return m.hram[addr-hramAddr]
	} else {
		panic(fmt.Sprintf("Unexpected memory address %#x", addr))
	}
}

// tick progresses the MMU by the given number of cycles.
func (m *mmu) tick(opTime int) {
	for i := 0; i < opTime; i++ {
		if m.dmaActive {
			// Work on a DMA transfer
			lower, _ := split16(m.dmaCursor)
			if lower < 0x9F {
				// Transfer a byte
				m.setNoNotify(oamRAMAddr+uint16(lower), m.at(m.dmaCursor))

				m.dmaCursor++
			}
			// Wait for the DMA transfer to finish
			if m.dmaCycleCount >= dmaCycleLength {
				m.dmaActive = false
			} else {
				m.dmaCycleCount++
			}
		}
	}
}

func (m *mmu) setNoNotify(addr uint16, val uint8) {
	if m.db != nil {
		m.db.memWriteHook(addr)
	}

	if addr < videoRAMAddr {
		// Somewhere in ROM, let the MBC handle it
		m.mbc.set(addr, val)
	} else if addr < bankedRAMAddr {
		// Video RAM
		m.videoRAM[addr-videoRAMAddr] = val
	} else if addr < ramAddr {
		// Banked RAM, let the MBC handle it
		m.mbc.set(addr, val)
	} else if addr < ramMirrorAddr {
		// General-purpose RAM
		m.ram[addr-ramAddr] = val
	} else if addr < oamRAMAddr {
		// A mirror to general-purpose RAM
		m.ram[addr-ramMirrorAddr] = val
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

func (m *mmu) set(addr uint16, val uint8) {
	// Notify any subscribers of this event
	if onWrite, ok := m.subscribers[addr]; ok {
		val = onWrite(addr, val)
	}

	m.setNoNotify(addr, val)
}

func (m *mmu) subscribeTo(addr uint16, onWrite onWriteFunc) {
	if _, ok := m.subscribers[addr]; ok {
		panic(fmt.Sprintf("attempt to have multiple subscribers to one address %#x", addr))
	} else {
		m.subscribers[addr] = onWrite
	}
}

func (m *mmu) dump() []uint8 {
	var data []uint8

	// Silence invalid read warnings and debugger
	oldPrintInstructions := printInstructions
	printInstructions = false
	m.db = nil

	for i := 0; i <= lastAddr; i++ {
		data = append(data, m.at(uint16(i)))
	}

	printInstructions = oldPrintInstructions

	return data
}

// onDMAWrite triggers when the special DMA address is written to. This
// triggers a DMA transfer, where data is copied into OAM RAM.
func (m *mmu) onDMAWrite(addr uint16, val uint8) uint8 {
	if m.dmaActive {
		// A DMA transfer is already active!
		// TODO(velovix): What is the hardware's behavior here?
		panic(fmt.Sprintf("DMA transfer request to %#x when "+
			"a transfer is already active", val))
	}
	// TODO(velovix): Lock all memory except HRAM?
	// Start a DMA transfer
	m.dmaActive = true
	// Use the value as the higher byte in the source address
	m.dmaCursor = uint16(val) << 8
	m.dmaCycleCount = 0

	return val
}

// mbc describes a memory bank controller.
type mbc interface {
	set(addr uint16, val uint8)
	at(addr uint16) uint8
}

// The start of each section of the memory map.
const (
	bank0ROMAddr     = 0x0000
	bankedROMAddr    = 0x4000
	videoRAMAddr     = 0x8000
	bankedRAMAddr    = 0xA000
	ramAddr          = 0xC000
	ramMirrorAddr    = 0xE000
	oamRAMAddr       = 0xFE00
	invalidArea2Addr = 0xFEA0
	ioAddr           = 0xFF00
	hramAddr         = 0xFF80
	lastAddr         = 0xFFFF
)
