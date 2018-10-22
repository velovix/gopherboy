package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"strings"
)

// noNullTerms removes null terminators from the given string.
func noNullTerms(str string) string {
	return strings.Trim(str, "\000")
}

func loadROM(env *environment, data io.Reader) error {
	cartridgeData, err := ioutil.ReadAll(data)
	if err != nil {
		return fmt.Errorf("loading cartridge data: %v", err)
	}

	fmt.Println("Cartridge Header:")

	// Read the cartridge title. Earlier games used 16 characters for this but
	// we're assuming 11 since that's what it is from CGB onwards
	title := noNullTerms(string(cartridgeData[0x0134:0x0140]))
	fmt.Println("  Title:", strings.Trim(title, "\000"))

	// Read the manufacturer code, a 4 character code. Not useful for anything.
	manufacturerCode := noNullTerms(string(cartridgeData[0x013F:0x0143]))
	fmt.Println("  Manufacturer Code:", manufacturerCode)

	// Read the CGB flag, which tells us if the game is compatible with the
	// Gameboy original
	cgbFlag := cartridgeData[0x0143]
	fmt.Println("  CGB Flag:", cgbFlag)
	if cgbFlag == 0xC0 {
		return fmt.Errorf("CGB-only roms are not supported")
	}

	// Read the licensee code, a two character ASCII code. Not actually useful
	// for us.
	licenseeCode := noNullTerms(string(cartridgeData[0x0144:0x0146]))
	fmt.Println("  Licensee Code:", licenseeCode)

	// Read the SGB flag, telling us if this ROM is SGB-compatible
	sgbFlag := cartridgeData[0x0146]
	fmt.Println("  SGB Flag:", sgbFlag)

	// Read the cartridge type. This describes what hardware is available in
	// the cartridge, like a memory bank switcher, battery, etc.
	cartridgeType := cartridgeData[0x0147]
	fmt.Println("  Cartridge Type:", cartridgeType)
	if cartridgeType != 0 {
		return fmt.Errorf("complicated cartridge types are not supported")
	}

	// Read the ROM size. This tells us how big the ROM is and consequently how
	// many banks are being used. This number is not the actual size of the
	// ROM, just a sentinel value.
	romSizeType := cartridgeData[0x0148]
	fmt.Println("  ROM Size Type:", romSizeType)
	if romSizeType != 0 {
		return fmt.Errorf("banking is not supported")
	}

	// Read the RAM size. This tells us how much additional RAM is included in
	// the cartridge. This number is not the actual RAM size, just a sentinel
	// value.
	ramSizeType := cartridgeData[0x0149]
	fmt.Println("  RAM Size Type:", ramSizeType)
	if ramSizeType != 0 {
		fmt.Println("cartridge RAM is not supported")
	}

	// Read the destination code. This tells us if the cartridge is for the
	// Japanese market or not.
	destinationCode := cartridgeData[0x014A]
	fmt.Println("  Destination Code:", destinationCode)

	// Read the old licensee code. This used to be the way that cartridges
	// would tell us who published the game, but future games used the new
	// licensee code instead.
	oldLicenseeCode := cartridgeData[0x014B]
	fmt.Println("  Old Licensee Code:", oldLicenseeCode)

	// Read the mask ROM version number. This is supposed to tell us the
	// "version" of the game, but is pretty much always 0 apparently.
	maskROMVersionNumber := cartridgeData[0x014C]
	fmt.Println("  Mask ROM Version Number:", maskROMVersionNumber)

	// The checksum of all previous header info
	headerChecksum := cartridgeData[0x014D]
	fmt.Println("  Header Checksum:", headerChecksum)

	// Put cartridge data into memory, up until the expected size. I've found
	// that ROMs can have a lot of padding at the end, so they don't end where
	// you would expect
	for i := 0; i < 0x3FFF; i++ {
		env.mem[i] = cartridgeData[i]
	}

	// 0x100 is the designated entry point of a Gameboy ROM
	env.regs[regPC].set(0x100)

	return nil
}
