package gameboy

// isUnmappedAddress is a slice whose index is a memory address and whose value
// is true if the address is unmapped. This array is filled in using the
// unmappedAddresses slice at program initialization.
var isUnmappedAddress [0x10000]bool

func init() {
	for _, addr := range unmappedAddresses {
		isUnmappedAddress[addr] = true
	}
}

// unmappedAddresses is a list of all unmapped addresses in the DMG.
var unmappedAddresses = []uint16{
	// Unused CGB registers
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
	// Misc unused addresses
	0xFF03,
	0xFF08,
	0xFF09,
	0xFF0A,
	0xFF0B,
	0xFF0C,
	0xFF0D,
	0xFF0E,
	0xFF15,
	0xFF1F,
	0xFF27,
	0xFF28,
	0xFF29,
	0xFF2A,
	0xFF2B,
	0xFF2C,
	0xFF2D,
	0xFF2E,
	0xFF2F,
	0xFF4C,
	0xFF4E,
	0xFF57,
	0xFF58,
	0xFF59,
	0xFF5A,
	0xFF5B,
	0xFF5C,
	0xFF5D,
	0xFF5E,
	0xFF5F,
	0xFF60,
	0xFF61,
	0xFF62,
	0xFF63,
	0xFF64,
	0xFF65,
	0xFF66,
	0xFF67,
	0xFF6C,
	0xFF6D,
	0xFF6E,
	0xFF6F,
	0xFF71,
	0xFF72,
	0xFF73,
	0xFF74,
	0xFF75,
	0xFF78,
	0xFF79,
	0xFF7A,
	0xFF7B,
	0xFF7C,
	0xFF7D,
	0xFF7E,
	0xFF7F,
}
