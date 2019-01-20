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

type onWriteFunc func(addr uint16, val uint8) uint8

func newMMU(bootROM []byte, cartridgeData []uint8, mbc mbc) *mmu {
	if len(bootROM) != bootROMEndAddr {
		panic(fmt.Sprintf("invalid boot ROM size %#x", len(bootROM)))
	}

	m := &mmu{
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

	// Unmapped areas of memory always read 0xFF
	if isUnmappedAddress(addr) {
		return 0xFF
	}

	switch {
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
	case inVideoRAMArea(addr):
		return m.videoRAM[addr-videoRAMAddr]
	case inRAMArea(addr):
		return m.ram[addr-ramAddr]
	case inBankedRAMArea(addr):
		// The MBC handles RAM banking and availability
		return m.mbc.at(addr)
	case inRAMMirrorArea(addr):
		// A bank 0 RAM mirror, forwarded to the MBC by masking it as a regular
		// bank 0 RAM access
		return m.mbc.at(addr - (ramMirrorAddr - ramAddr))
	case inOAMArea(addr):
		return m.oamRAM[addr-oamRAMAddr]
	case inInvalidArea(addr):
		// Invalid area, which always returns 0xFF since it's the MMU's default
		// value
		if printInstructions {
			fmt.Printf("Warning: Read from invalid memory address %#x\n", addr)
		}
		return 0xFF
	case inIOArea(addr):
		return m.ioRAM[addr-ioAddr]
	case inHRAMArea(addr):
		return m.hram[addr-hramAddr]
	default:
		panic(fmt.Sprintf("Unexpected memory address %#x", addr))
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
	if isUnmappedAddress(addr) {
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
		m.db.memWriteHook(addr)
	}

	switch {
	case inBank0ROMArea(addr) || inBankedROMArea(addr):
		// "Writes" to ROM areas are used to control MBCs
		m.mbc.set(addr, val)
	case inVideoRAMArea(addr):
		m.videoRAM[addr-videoRAMAddr] = val
	case inRAMArea(addr):
		m.ram[addr-ramAddr] = val
	case inBankedRAMArea(addr):
		// The MBC handles RAM banking and availability
		m.mbc.set(addr, val)
	case inRAMMirrorArea(addr):
		// An area that mirrors built-in RAM. Forward it to the MBC disguised
		// as a regular write
		m.mbc.set(addr-(ramMirrorAddr-ramAddr), val)
	case inOAMArea(addr):
		m.oamRAM[addr-oamRAMAddr] = val
	case inInvalidArea(addr):
		if printInstructions {
			fmt.Printf("Warning: Write to invalid area %#x", addr)
		}
	case inIOArea(addr):
		m.ioRAM[addr-ioAddr] = val
	case inHRAMArea(addr):
		m.hram[addr-hramAddr] = val
	default:
		panic(fmt.Sprintf("Unexpected memory address %#x", addr))
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

// unusedCGBRegisters are all registers that are used on the CGB but unused on the DMG
var unusedCGBRegisters = []uint16{
	key1Addr,
	vbkAddr,
	hdma1Addr,
	hdma2Addr,
	hdma3Addr,
	hdma4Addr,
	hdma5Addr,
	rpAddr,
	bcpsAddr,
	bcpdAddr,
	ocpsAddr,
	ocpdAddr,
	svbkAddr,
	pcm12Ch2Addr,
	pcm34Ch4Addr,
}

// isUnmappedAddress returns true if the given address is in an unmapped range
// of memory. These areas cannot be written to and always read 0xFF.
func isUnmappedAddress(addr uint16) bool {
	if addr == 0xFF03 {
		return true
	}
	if addr >= 0xFF08 && addr <= 0xFF0E {
		return true
	}
	if addr == 0xFF15 {
		return true
	}
	if addr == 0xFF1F {
		return true
	}
	if addr >= 0xFF27 && addr <= 0xFF2F {
		return true
	}
	if addr == 0xFF4C {
		return true
	}
	if addr == 0xFF4E {
		return true
	}
	if addr >= 0xFF57 && addr <= 0xFF67 {
		return true
	}
	if addr >= 0xFF6C && addr <= 0xFF6F {
		return true
	}
	if addr == 0xFF71 {
		return true
	}
	if addr >= 0xFF72 && addr <= 0xFF75 {
		return true
	}
	if addr >= 0xFF78 && addr <= 0xFF7F {
		return true
	}

	for _, cgbReg := range unusedCGBRegisters {
		if addr == cgbReg {
			return true
		}
	}

	return false
}
