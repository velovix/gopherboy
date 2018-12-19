package gameboy

import (
	"bytes"
	"fmt"
	"strings"
)

type romHeader struct {
	// The title of the game.
	title string
	// Some four character code with an unknown meaning.
	manufacturerCode string
	// Tells us if the game requires the CGB or can run on the original DMG
	// Game Boy.
	cgbFlag uint8
	// A two character code indicating the game's publisher.
	licenseeCode string
	// Tells us if the game is SGB-compatible.
	sgbFlag uint8
	// This describes what hardware is available in the cartridge, like a memory
	// bank switcher, battery, etc.
	cartridgeType uint8
	// This tells us how big the ROM is and consequently how many banks are
	// being used. This number is not the actual size of the ROM, just a
	// sentinel value.
	romSizeType uint8
	// This tells us how much additional RAM is included in the cartridge. This
	// number is not the actual RAM size, just a sentinel value.
	ramSizeType uint8
	// This tells us if the cartridge is for the Japanese market or not.
	destinationCode uint8
	// This used to be the way that cartridges would tell us who published the
	// game, but future games used the new licensee code instead.
	oldLicenseeCode uint8
	// This is supposed to tell us the "version" of the game, but is pretty
	// much always 0 apparently.
	maskROMVersionNumber uint8
	// The checksum of all previous header info
	headerChecksum uint8
}

func loadROMHeader(cartridgeData []byte) romHeader {
	return romHeader{
		title:                noNullTerms(string(cartridgeData[0x0134:0x0140])),
		manufacturerCode:     noNullTerms(string(cartridgeData[0x013F:0x0143])),
		cgbFlag:              cartridgeData[0x0143],
		licenseeCode:         noNullTerms(string(cartridgeData[0x0144:0x0146])),
		sgbFlag:              cartridgeData[0x0146],
		cartridgeType:        cartridgeData[0x0147],
		romSizeType:          cartridgeData[0x0148],
		ramSizeType:          cartridgeData[0x0149],
		destinationCode:      cartridgeData[0x014A],
		oldLicenseeCode:      cartridgeData[0x014B],
		maskROMVersionNumber: cartridgeData[0x014C],
		headerChecksum:       cartridgeData[0x014D],
	}

}

// noNullTerms removes null terminators from the given string.
func noNullTerms(str string) string {
	return strings.Trim(str, "\000")
}

func (h romHeader) String() string {
	str := bytes.NewBufferString("")

	fmt.Fprintln(str, "Cartridge Header:")
	fmt.Fprintln(str, "  Title:", h.title)
	fmt.Fprintln(str, "  Manufacturer Code:", h.manufacturerCode)
	fmt.Fprintln(str, "  CGB Flag:", h.cgbFlag)
	fmt.Fprintln(str, "  Licensee Code:", h.licenseeCode)
	fmt.Fprintln(str, "  SGB Flag:", h.sgbFlag)
	fmt.Fprintln(str, "  Cartridge Type:", h.cartridgeType)
	fmt.Fprintln(str, "  ROM Size Type:", h.romSizeType)
	fmt.Fprintln(str, "  RAM Size Type:", h.ramSizeType)
	fmt.Fprintln(str, "  Destination Code:", h.destinationCode)
	fmt.Fprintln(str, "  Old Licensee Code:", h.oldLicenseeCode)
	fmt.Fprintln(str, "  Mask ROM Version Number:", h.maskROMVersionNumber)
	fmt.Fprintln(str, "  Header Checksum:", h.headerChecksum)

	return str.String()
}
