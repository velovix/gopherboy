package main

const sampleRate = 44100

type soundController struct {
	sinceLastSample int

	pulseATrigger   bool
	pulseAUseLength bool
	pulseALength    uint8
	pulseADutyCycle float64

	env *environment
}

func newSoundController(env *environment) *soundController {
	return &soundController{
		sinceLastSample: cpuClockRate / sampleRate,
		env:             env,
	}
}

func (sc *soundController) tick(cycles int) {
	sc.sinceLastSample += cycles

	if sc.sinceLastSample >= cpuClockRate/sampleRate {

	}
}
