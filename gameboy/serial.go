package gameboy

// serial controls transfer operations to and from the Game Boy through the
// link cable.
type serial struct{}

func newSerial(state *State) *serial {
	s := &serial{}

	state.mmu.subscribeTo(scAddr, s.onSCWrite)

	return s
}

// onSCWrite triggers when the SIO Control register is written to. It
// configures or starts a serial transfer.
func (s *serial) onSCWrite(addr uint16, val uint8) uint8 {
	// TODO(velovix): Actually trigger some behavior

	// Unused bits 6-1 are always 1
	return val | 0x7E
}
