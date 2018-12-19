package gameboy

const sampleRate = 44100

type soundController struct {
	sinceLastSample int

	pulseATrigger   bool
	pulseAUseLength bool
	pulseALength    uint8
	pulseADutyCycle float64

	state *State
}

func newSoundController(state *State) *soundController {
	return &soundController{
		sinceLastSample: cpuClockRate / sampleRate,
		state:           state,
	}
}

func (sc *soundController) tick(cycles int) {
	sc.sinceLastSample += cycles

	if sc.sinceLastSample >= cpuClockRate/sampleRate {

	}
}
