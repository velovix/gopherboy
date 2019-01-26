package gameboy

import "fmt"

// makeROMBanks creates the necessary amount of ROM banks as specified by the
// given ROM size type, then returns it as a map whose key is a ROM bank number
// and whose value is the corresponding ROM bank.
func makeROMBanks(romSizeType uint8, cartridgeData []uint8) [][]uint8 {
	var romBankCount int

	switch romSizeType {
	case 0x00:
		// A single ROM bank, no switching going on
		romBankCount = 2
	case 0x01:
		// Four ROM banks, 16 KB in size
		romBankCount = 4
	case 0x02:
		// Eight ROM banks, 16 KB in size
		romBankCount = 8
	case 0x03:
		// Sixteen ROM banks, 16 KB in size
		romBankCount = 16
	case 0x04:
		// Thirty-two ROM banks, 16 KB in size
		romBankCount = 32
	case 0x05:
		// Sixty-four ROM banks, 16 KB in size
		romBankCount = 64
	case 0x06:
		// 128 ROM banks, 16 KB in size
		romBankCount = 128
	case 0x07:
		// 256 ROM banks, 16 KB in size
		romBankCount = 256
	case 0x08:
		// 512 ROM banks, 16 KB in size
		romBankCount = 512
	default:
		panic(fmt.Sprintf("Unsupported ROM size type %v", romSizeType))
	}

	romBanks := make([][]uint8, romBankCount)

	// Create the ROM banks
	for i := 0; i < romBankCount; i++ {
		romBanks[i] = make([]uint8, 0x4000)
	}

	// Put cartridge data into each bank
	for bank, data := range romBanks {
		startAddr := 0x4000 * int(bank)
		for i := startAddr; i < startAddr+0x4000; i++ {
			data[i-startAddr] = cartridgeData[i]
		}
	}

	return romBanks
}

// makeRAMBanks creates the necessary amount of external RAM banks as specified
// by the given RAM size type, then returns it as a map whose key is a RAM bank
// number and whose value is the corresponding RAM bank.
func makeRAMBanks(ramSizeType uint8) (ramBanks [][]uint8) {
	switch ramSizeType {
	case 0x00:
		// No bank
	case 0x01:
		// One 2 KB bank
		ramBanks = append(ramBanks, make([]uint8, 2000))
	case 0x02:
		// One 8 KB bank
		ramBanks = append(ramBanks, make([]uint8, 0x2000))
	case 0x03:
		// Four 8 KB banks
		for i := 0; i < 4; i++ {
			ramBanks = append(ramBanks, make([]uint8, 0x2000))
		}
	case 0x04:
		// Sixteen 8 KB banks
		for i := 0; i < 16; i++ {
			ramBanks = append(ramBanks, make([]uint8, 0x2000))
		}
	case 0x05:
		// Eight 8 KB banks
		for i := 0; i < 8; i++ {
			ramBanks = append(ramBanks, make([]uint8, 0x2000))
		}
	default:
		panic(fmt.Sprintf("Unknown RAM size type %v", ramSizeType))
	}

	return ramBanks
}

func inBootROMArea(addr uint16) bool {
	return addr < bootROMEndAddr
}

func inBank0ROMArea(addr uint16) bool {
	return addr < bankedROMAddr
}

func inBankedROMArea(addr uint16) bool {
	return addr >= bankedROMAddr && addr < videoRAMAddr
}

func inVideoRAMArea(addr uint16) bool {
	return addr >= videoRAMAddr && addr < bankedRAMAddr
}

func inBankedRAMArea(addr uint16) bool {
	return addr >= bankedRAMAddr && addr < ramAddr
}

func inRAMArea(addr uint16) bool {
	return addr >= ramAddr && addr < ramMirrorAddr
}

func inRAMMirrorArea(addr uint16) bool {
	return addr >= ramMirrorAddr && addr < oamRAMAddr
}

func inOAMArea(addr uint16) bool {
	return addr >= oamRAMAddr && addr < invalidArea2Addr
}

func inInvalidArea(addr uint16) bool {
	return addr >= invalidArea2Addr && addr < ioAddr
}

func inIOArea(addr uint16) bool {
	return addr >= ioAddr && addr < hramAddr
}

func inHRAMArea(addr uint16) bool {
	return addr >= hramAddr
}

// The start of each section of the memory map.
const (
	bootROMAddr    = 0x0000
	bootROMEndAddr = 0x0100
	bank0ROMAddr   = 0x0000
	bankedROMAddr  = 0x4000
	videoRAMAddr   = 0x8000
	bankedRAMAddr  = 0xA000
	ramAddr        = 0xC000
	ramMirrorAddr  = 0xE000
	oamRAMAddr     = 0xFE00
	// TODO(velovix): Rename this since there's only one invalid area
	invalidArea2Addr = 0xFEA0
	ioAddr           = 0xFF00
	hramAddr         = 0xFF80
	lastAddr         = 0xFFFF
)
