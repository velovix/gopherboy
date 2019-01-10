package gameboy

const sampleRate = 44100

// soundController emulates the Game Boy's sound chip. It produces audio data
// that may be played.
type soundController struct {
	state *State
}

func newSoundController(state *State) *soundController {
	sc := &soundController{
		state: state,
	}

	sc.state.mmu.subscribeTo(nr10Addr, sc.onNR10Write)
	sc.state.mmu.subscribeTo(nr30Addr, sc.onNR30Write)
	sc.state.mmu.subscribeTo(nr32Addr, sc.onNR32Write)

	return sc
}

func (sc *soundController) tick(cycles int) {
	// TODO(velovix): Implement this. Biggest TODO ever
}

// onNR10Write is called when the Sound Mode 1 Sweep register is written to.
func (sc *soundController) onNR10Write(addr uint16, val uint8) uint8 {
	// Bit 7 is unused and always 1
	return val | 0x80
}

// onNR30Write is called when the Sound Mode 3 On/Off register is written to.
func (sc *soundController) onNR30Write(addr uint16, val uint8) uint8 {
	// Bits 6-0 are unused and always 1
	return val | 0x7F
}

// onNR32Write is called when the Sound Mode 3 Select Output Level register is
// written to.
func (sc *soundController) onNR32Write(addr uint16, val uint8) uint8 {
	// Bits 7 and bits 4-0 are unused and always 1
	return val | 0x9F
}
