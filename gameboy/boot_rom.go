package gameboy

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io/ioutil"
)

// BootROM returns a modified version of the DMG boot ROM with the original
// logo replaced with misc/boot_rom_logo.png to avoid Copyright concerns.
func BootROM() []byte {
	decoder := base64.NewDecoder(base64.StdEncoding, bytes.NewBuffer([]byte(bootROM)))
	data, err := ioutil.ReadAll(decoder)
	if err != nil {
		panic(fmt.Errorf("while decoding boot ROM: %v", err))
	}

	return data
}

const bootROM = "Mf7/ryH/nzLLfCD7ISb/DhE+gDLiDD7z4jI+d3c+/OBHEagAIRCAGs2VAM2WABN7/jQg8xHYAAYI" +
	"GhMiIwUg+T4Z6hCZIS+ZDgw9KAgyDSD5Lg8Y82c+ZFfgQj6R4EAEHgIODPBE/pAg+g0g9x0g8g4T" +
	"JHweg/5iKAYewf5kIAZ74gw+h+LwQpDgQhUg0gUgTxYgGMtPBgTFyxEXwcsRFwUg9SIjIiPJAAAA" +
	"AAAAaa8Dw/AA8AAMPGlfAAAAAAAAAAAAAAERyCMQiIxEEyKAETFMCIgAAAAAAAAAAAAAAAAhBAER" +
	"qAAaE74gASN9/jQg9QYZeIYjBSD7hiABPgHgUA=="
