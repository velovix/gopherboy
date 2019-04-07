package gameboy

import (
	"fmt"
)

// dmaCycleLength is the number of cycles a DMA transfer takes.
const dmaCycleLength = 671

// mmu is the memory management unit. It handles all operations that are common
// to all Game Boy games and defers to the cartridge's memory bank controller
// in cartridge-specific cases.
type mmu struct {
	memory *[0x10000]uint8

	// bootROM refers to the first 256 bytes of ROM, which is where the Game
	// Boy boot sequence is stored.
	bootROM []uint8
	// ram is the built-in RAM on the device.
	ram []uint8
	// videoRAM is where sprite and tile data is stored for the video
	// controller to access.
	videoRAM []uint8
	// oamRAM is where sprite attribute data is stored for the video controller
	// to access.
	oamRAM []uint8
	// ioRAM is where memory registers are mapped to for controlling and
	// reading various aspects of the hardware.
	ioRAM []uint8
	// hram is a special general-purpose RAM area
	hram []uint8

	// bootROMEnabled is true if the boot ROM is available at 0x0000 to 0x0100.
	// The boot ROM itself turns this off before the game starts. If this is
	// turned off, 0x0000 to 0x0100 maps to ROM bank 0.
	bootROMEnabled bool
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

// mbc describes a memory bank controller.
type mbc interface {
	set(addr uint16, val uint8)
	at(addr uint16) uint8
}

// batteryBackedMBC is a memory bank controller with RAM that is
// battery-backed, meaning that it can be saved after the device
// is powered off.
type batteryBackedMBC interface {
	mbc
	// dumpBatteryBackedRAM returns a dump of all RAM that is battery-backed in
	// the MBC.
	dumpBatteryBackedRAM() []uint8
	// loadBatteryBackedRAM loads the given data into all battery-backed RAM in
	// the MBC.
	loadBatteryBackedRAM(dump []uint8)
}

type onWriteFunc func(addr uint16, val uint8) uint8

func newMMU(bootROM []byte, cartridgeData []uint8, mbc mbc) *mmu {
	if len(bootROM) != bootROMEndAddr {
		panic(fmt.Sprintf("invalid boot ROM size %#x", len(bootROM)))
	}

	m := &mmu{
		memory:         &[0x10000]uint8{},
		bootROM:        bootROM,
		ram:            make([]uint8, ramMirrorAddr-ramAddr),
		bootROMEnabled: true,
		videoRAM:       make([]uint8, bankedRAMAddr-videoRAMAddr),
		oamRAM:         make([]uint8, invalidArea2Addr-oamRAMAddr),
		ioRAM:          make([]uint8, hramAddr-ioAddr),
		hram:           make([]uint8, lastAddr-hramAddr+1),
		subscribers:    make(map[uint16]onWriteFunc),
		mbc:            mbc,
	}

	m.subscribeTo(bootROMDisableAddr, m.onBootROMDisableWrite)
	m.subscribeTo(dmaAddr, m.onDMAWrite)

	return m
}

// at returns the value in the given address.
func (m *mmu) at(addr uint16) uint8 {
	if m.db != nil {
		m.db.memReadHook(addr)
	}

	switch {
	case isUnmappedAddress[addr]:
		// Unmapped areas of memory always read 0xFF
		return 0xFF
	case inBootROMArea(addr):
		// Either the boot ROM if it's enabled, or a ROM access
		if m.bootROMEnabled {
			return m.bootROM[addr-bootROMAddr]
		} else {
			return m.mbc.at(addr)
		}
	case inBank0ROMArea(addr):
		return m.mbc.at(addr)
	case inBankedROMArea(addr):
		// Some additional ROM bank, controlled by the MBC
		return m.mbc.at(addr)
	case inBankedRAMArea(addr):
		// The MBC handles RAM banking and availability
		return m.mbc.at(addr)
	case inRAMMirrorArea(addr):
		// A bank 0 RAM mirror
		return m.memory[addr-(ramMirrorAddr-ramAddr)]
	case inInvalidArea(addr):
		// Invalid area, which always returns 0xFF since it's the MMU's default
		// value
		if printWarnings {
			fmt.Printf("Warning: Read from invalid memory address %#x\n", addr)
		}
		return 0xFF
	default:
		return m.memory[addr]
	}
}

// tick progresses the MMU by the given number of cycles.
func (m *mmu) tick(opTime int) {
	for i := 0; i < opTime; i++ {
		if m.dmaActive {
			// Work on a DMA transfer
			lower, _ := split16(m.dmaCursor)
			if lower <= 0x9F {
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

// set requests the MMU to set the value at the given address to the given
// value. This method notifies any subscribed devices about this write, meaning
// that side effects may occur.
func (m *mmu) set(addr uint16, val uint8) {
	// Unmapped addresses cannot be written to
	if isUnmappedAddress[addr] {
		return
	}

	// Notify any subscribers of this event
	if onWrite, ok := m.subscribers[addr]; ok {
		val = onWrite(addr, val)
	}

	m.setNoNotify(addr, val)
}

// setNoNotify requests the MMU to set the value at the given address to the
// given value. Subscribed devices are not notified. This is useful for devices
// that might incorrectly trigger themselves when writing to a place in memory.
func (m *mmu) setNoNotify(addr uint16, val uint8) {
	if m.db != nil {
		m.db.memWriteHook(addr, val)
	}

	switch {
	case inBank0ROMArea(addr) || inBankedROMArea(addr):
		// "Writes" to ROM areas are used to control MBCs
		m.mbc.set(addr, val)
	case inBankedRAMArea(addr):
		// The MBC handles RAM banking and availability
		m.mbc.set(addr, val)
	case inInvalidArea(addr):
		if printWarnings {
			fmt.Printf("Warning: Write to invalid area %#x\n", addr)
		}
	default:
		m.memory[addr] = val
	}
}

// subscribeTo sets up the given function to be called when a value is written
// to the given address.
func (m *mmu) subscribeTo(addr uint16, onWrite onWriteFunc) {
	if _, ok := m.subscribers[addr]; ok {
		panic(fmt.Sprintf("attempt to have multiple subscribers to one address %#x", addr))
	} else {
		m.subscribers[addr] = onWrite
	}
}

// onDMAWrite triggers when the special DMA address is written to. This
// triggers a DMA transfer, where data is copied into OAM RAM.
func (m *mmu) onDMAWrite(addr uint16, val uint8) uint8 {
	// Nothing happens if a DMA transfer is already happening
	if !m.dmaActive {
		// TODO(velovix): Lock all memory except HRAM?
		// Start a DMA transfer
		m.dmaActive = true
		// Use the value as the higher byte in the source address
		m.dmaCursor = uint16(val) << 8
		m.dmaCycleCount = 0
	}

	return val
}

// onBootROMDisableWrite triggers when the boot ROM disable register is written
// to. It disables the boot ROM.
func (m *mmu) onBootROMDisableWrite(addr uint16, val uint8) uint8 {
	if m.bootROMEnabled {
		fmt.Println("Disabled boot ROM")
		m.bootROMEnabled = false
	}

	// This register always reads 0xFF
	return 0xFF
}
