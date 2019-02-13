package gameboy

import (
	"fmt"
	"math"
)

const (
	// frameSequencerClockRate is the clock rate of the frame sequencer, a
	// clock used to time sound operations.
	frameSequencerClockRate = 512
	// durationClockRate is the clock rate of the duration clock. When this
	// ticks, it decreases the duration counter of a voice by 1. This is
	// clocked by the frame sequencer.
	durationClockRate = 256
	// volumeClockRate is the clock rate of the volume clock. When this ticks,
	// it decreases or increases the volume of a voice by 1. This is clocked by
	// the frame sequencer.
	volumeClockRate = 64
	// frequencyClockRate is the clock rate of the frequency clock. When this
	// ticks, it decreases or increases the frequency of Pulse A by some
	// amount. This is clocked by the frame sequencer.
	frequencyClockRate = 128
)

// SoundController emulates the Game Boy's sound chip. It produces audio data
// that may be played.
type SoundController struct {
	state *State

	// A clock that runs at 512 Hz. It is used to time sound operations.
	frameSequencer int
	// A clock that increments with every CPU cycle. Used by the frame
	// sequencer.
	tClock int

	// If false, the whole controller goes to sleep and now sound is emitted
	Enabled bool
	// Volume for the left and right audio channels.
	leftVolume  int
	rightVolume int

	PulseA *PulseA
	PulseB *PulseB
	Wave   *Wave
	Noise  *Noise
}

type PulseA struct {
	On           bool
	RightEnabled bool
	LeftEnabled  bool

	volume    int
	frequency int

	duration    int
	useDuration bool

	volumePeriod int
	amplify      bool

	dutyCycle int

	lastFrequency     int
	frequencyPeriod   int
	attenuate         bool
	sweepShift        uint
	useFrequencySweep bool
}

func (voice *PulseA) tick(frameSequencer int) {
	if !voice.On {
		return
	}

	// Decrease the duration if the duration clock ticked
	if voice.useDuration && frameSequencer%(frameSequencerClockRate/durationClockRate) == 0 {
		voice.duration--
		if voice.duration == 0 {
			voice.On = false
		}
	}

	// Sweep the volume up or down
	if voice.volumePeriod != 0 {
		// Calculate the resulting volume clock rate from the base clock rate
		// and the configurable period value
		clockRate := volumeClockRate / voice.volumePeriod
		// Increment or decrement the volume on the volume clock's tick if the
		// volume isn't already at max or min
		if frameSequencer%(frameSequencerClockRate/clockRate) == 0 {
			if voice.amplify && voice.volume < 15 {
				voice.volume++
			} else if voice.volume > 0 {
				voice.volume--
			}
		}
	}

	if voice.useFrequencySweep && voice.frequencyPeriod != 0 {
		// Calculate the resulting frequency clock rate from the base clock
		// rate and the configurable period value
		clockRate := frequencyClockRate / voice.frequencyPeriod
		// Increment or decrement the frequency on the frequency clock's tick
		// if the frequency isn't already at max or min
		if frameSequencer%(frameSequencerClockRate/clockRate) == 0 {
			// Calculate the frequency step by shifting the initial frequency of
			// the voice by the configured amount
			step := voice.lastFrequency >> voice.sweepShift

			newFrequency := voice.lastFrequency
			if voice.attenuate {
				newFrequency -= step
			} else {
				newFrequency += step
			}

			// Check for an overflow. The frequency is an 11-bit value
			if newFrequency > 2047 || newFrequency < 0 {
				voice.useFrequencySweep = false
			} else {
				voice.lastFrequency = newFrequency
				voice.frequency = newFrequency

				// Check for a future overflow. I know this seems weird but
				// this is apparently how the hardware does it
				if voice.attenuate {
					newFrequency -= step
				} else {
					newFrequency += step
				}
				if newFrequency > 2047 || newFrequency < 0 {
					voice.useFrequencySweep = false
				}
			}
		}
	}
}

func (voice *PulseA) Volume() float64 {
	return float64(voice.volume) / 15.0
}

func (voice *PulseA) Frequency() float64 {
	return 131072.0 / (2048.0 - float64(voice.frequency))
}

func (voice *PulseA) DutyCycle() float64 {
	switch voice.dutyCycle {
	case 0:
		return 0.125
	case 1:
		return 0.25
	case 2:
		return 0.50
	case 3:
		return 0.75
	default:
		panic(fmt.Sprintf("invalid duty cycle value %v", voice.dutyCycle))
	}
}

type PulseB struct {
	On           bool
	RightEnabled bool
	LeftEnabled  bool

	volume    int
	frequency int

	duration    int
	useDuration bool

	volumePeriod int
	amplify      bool

	dutyCycle int
}

func (voice *PulseB) tick(frameSequencer int) {
	if !voice.On {
		return
	}

	// Decrease the duration if the duration clock ticked
	if voice.useDuration && frameSequencer%(frameSequencerClockRate/durationClockRate) == 0 {
		voice.duration--
		if voice.duration == 0 {
			voice.On = false
		}
	}

	// Sweep the volume up or down
	if voice.volumePeriod != 0 {
		// Calculate the resulting volume clock rate from the base clock rate
		// and the configurable period value
		clockRate := volumeClockRate / voice.volumePeriod
		// Increment or decrement the volume on the volume clock's tick if the
		// volume isn't already at max or min
		if frameSequencer%(frameSequencerClockRate/clockRate) == 0 {
			if voice.amplify && voice.volume < 15 {
				voice.volume++
			} else if voice.volume > 0 {
				voice.volume--
			}
		}
	}
}

func (voice *PulseB) Volume() float64 {
	return float64(voice.volume) / 15.0
}

func (voice *PulseB) Frequency() float64 {
	return 131072.0 / (2048.0 - float64(voice.frequency))
}

func (voice *PulseB) DutyCycle() float64 {
	switch voice.dutyCycle {
	case 0:
		return 0.125
	case 1:
		return 0.25
	case 2:
		return 0.50
	case 3:
		return 0.75
	default:
		panic(fmt.Sprintf("invalid duty cycle value %v", voice.dutyCycle))
	}
}

type Wave struct {
	On           bool
	RightEnabled bool
	LeftEnabled  bool

	volume    int
	frequency int
	pattern   []uint8

	duration    int
	useDuration bool

	rightShiftCode int
}

func (voice *Wave) tick(frameSequencer int) {
	if !voice.On {
		return
	}

	// Decrease the duration if the duration clock ticked
	if voice.useDuration && frameSequencer%(frameSequencerClockRate/durationClockRate) == 0 {
		voice.duration--
		if voice.duration == 0 {
			voice.On = false
		}
	}
}

func (voice *Wave) Volume() float64 {
	return float64(voice.volume)
}

func (voice *Wave) Frequency() float64 {
	return 65536 / (2048 - float64(voice.frequency))
}

func (voice *Wave) Pattern() [32]float64 {
	var pattern [32]float64

	var rightShift uint
	switch voice.rightShiftCode {
	case 0:
		rightShift = 4
	case 1:
		rightShift = 0
	case 2:
		rightShift = 1
	case 3:
		rightShift = 2
	}

	for i := 0; i < len(voice.pattern); i++ {
		value := voice.pattern[i] >> rightShift
		pattern[i] = float64(value) / 15.0
	}

	return pattern
}

type Noise struct {
	On           bool
	LeftEnabled  bool
	RightEnabled bool

	duration    int
	useDuration bool

	volume       int
	volumePeriod int
	amplify      bool

	shiftClockFrequency int
	dividingRatio       int
	lfsr                uint16
	widthMode           LFSRWidthMode
}

func (voice *Noise) tick(frameSequencer int) {
	if !voice.On {
		return
	}

	// Decrease the duration if the duration clock ticked
	if voice.useDuration && frameSequencer%(frameSequencerClockRate/durationClockRate) == 0 {
		voice.duration--
		if voice.duration == 0 {
			voice.On = false
		}
	}

	// Sweep the volume up or down
	if voice.volumePeriod != 0 {
		// Calculate the resulting volume clock rate from the base clock rate
		// and the configurable period value
		clockRate := volumeClockRate / voice.volumePeriod
		// Increment or decrement the volume on the volume clock's tick if the
		// volume isn't already at max or min
		if frameSequencer%(frameSequencerClockRate/clockRate) == 0 {
			if voice.amplify && voice.volume < 15 {
				voice.volume++
			} else if voice.volume > 0 {
				voice.volume--
			}
		}
	}
}

func (voice *Noise) ShiftFrequency() float64 {
	var dividingRatio float64
	if voice.dividingRatio == 0 {
		dividingRatio = 0.5
	} else {
		dividingRatio = float64(voice.dividingRatio)
	}

	return 524288.0 /
		dividingRatio /
		math.Pow(2, float64(voice.shiftClockFrequency+1))
}

func (voice *Noise) WidthMode() LFSRWidthMode {
	return voice.widthMode
}

func (voice *Noise) Volume() float64 {
	return float64(voice.volume) / 15
}

type LFSRWidthMode int

const (
	WidthMode7Bit  LFSRWidthMode = 7
	WidthMode15Bit LFSRWidthMode = 15
)

func newSoundController(state *State) *SoundController {
	sc := &SoundController{
		state:  state,
		PulseA: &PulseA{},
		PulseB: &PulseB{},
		Wave:   &Wave{},
		Noise:  &Noise{},
	}

	sc.state.mmu.subscribeTo(nr10Addr, sc.onNR10Write)
	sc.state.mmu.subscribeTo(nr14Addr, sc.onNR14Write)
	sc.state.mmu.subscribeTo(nr24Addr, sc.onNR24Write)
	sc.state.mmu.subscribeTo(nr30Addr, sc.onNR30Write)
	sc.state.mmu.subscribeTo(nr32Addr, sc.onNR32Write)
	sc.state.mmu.subscribeTo(nr34Addr, sc.onNR34Write)
	sc.state.mmu.subscribeTo(nr41Addr, sc.onNR41Write)
	sc.state.mmu.subscribeTo(nr44Addr, sc.onNR44Write)
	sc.state.mmu.subscribeTo(nr50Addr, sc.onNR50Write)
	sc.state.mmu.subscribeTo(nr51Addr, sc.onNR51Write)
	sc.state.mmu.subscribeTo(nr52Addr, sc.onNR52Write)

	return sc
}

// LeftVolume returns the global volume control of the left channel, from 0 to
// 1.
func (sc *SoundController) LeftVolume() float64 {
	return float64(sc.leftVolume) / 7
}

// RightVolume returns the global volume control of the right channel, from 0
// to 1.
func (sc *SoundController) RightVolume() float64 {
	return float64(sc.rightVolume) / 7
}

func (sc *SoundController) tick(cycles int) {
	for i := 0; i < cycles; i++ {
		// Update the frame sequencer
		sc.tClock++
		if sc.tClock%(cpuClockRate/frameSequencerClockRate) == 0 {
			sc.frameSequencer++
			sc.PulseA.tick(sc.frameSequencer)
			sc.PulseB.tick(sc.frameSequencer)
			sc.Wave.tick(sc.frameSequencer)
			sc.Noise.tick(sc.frameSequencer)
		}
		if sc.tClock == cpuClockRate {
			sc.tClock = 0
		}

	}
}

// onNR10Write is called when the Sound Mode 1 Sweep register is written to.
func (sc *SoundController) onNR10Write(addr uint16, val uint8) uint8 {
	// Bit 7 is unused and always 1
	return val | 0x80
}

// onNR14Write is called when the NR14 memory register is written to. When a 1
// is written to bit 7 of this register, the Pulse A voice is restarted with
// the configuration found in this register and others.
func (sc *SoundController) onNR14Write(addr uint16, val uint8) uint8 {
	if val&0x80 == 0x80 {
		sc.PulseA.On = true

		// Load duration information
		duration := sc.state.mmu.at(nr11Addr) & 0x3F
		sc.PulseA.duration = 64 - int(duration)
		sc.PulseA.useDuration = val&0x40 == 0x40

		// Load volume and volume sweep information
		nr12 := sc.state.mmu.at(nr12Addr)
		volume := (nr12 & 0xF0) >> 4
		volumePeriod := nr12 & 0x7
		amplify := nr12&0x8 == 0x8

		dutyCycle := (sc.state.mmu.at(nr11Addr) & 0xC0) >> 6

		// Load frequency and frequency sweep information
		frequency := uint16(sc.state.mmu.at(nr13Addr))
		frequency |= uint16(val&0x7) << 8
		nr10 := sc.state.mmu.at(nr10Addr)
		frequencyPeriod := (nr10 & 0x70) >> 4
		attenuate := nr10&0x08 == 0x08
		sweepShift := nr10 & 0x07

		sc.PulseA.volume = int(volume)
		sc.PulseA.volumePeriod = int(volumePeriod)
		sc.PulseA.amplify = amplify

		sc.PulseA.dutyCycle = int(dutyCycle)

		sc.PulseA.frequency = int(frequency)
		sc.PulseA.lastFrequency = int(frequency)
		sc.PulseA.frequencyPeriod = int(frequencyPeriod)
		sc.PulseA.attenuate = attenuate
		sc.PulseA.sweepShift = uint(sweepShift)
		sc.PulseA.useFrequencySweep = sweepShift != 0 || frequencyPeriod != 0
	}

	return val
}

// onNR24Write is called when the NR24 memory register is written to. When a 1
// is written to bit 7 of this register, the Pulse B voice is restarted with
// the configuration found in this register and others.
func (sc *SoundController) onNR24Write(addr uint16, val uint8) uint8 {
	if val&0x80 == 0x80 {
		sc.PulseB.On = true

		// Load duration information
		duration := sc.state.mmu.at(nr21Addr) & 0x3F
		sc.PulseB.duration = 64 - int(duration)
		sc.PulseB.useDuration = val&0x40 == 0x40

		// Load frequency information
		frequency := uint16(sc.state.mmu.at(nr23Addr))
		frequency |= uint16(val&0x7) << 8

		// Load volume and volume sweep information
		nr22 := sc.state.mmu.at(nr22Addr)
		volume := (nr22 & 0xF0) >> 4
		amplify := nr22&0x8 == 0x8
		volumePeriod := nr22 & 0x7

		dutyCycle := (sc.state.mmu.at(nr21Addr) & 0xC0) >> 6

		sc.PulseB.volume = int(volume)
		sc.PulseB.volumePeriod = int(volumePeriod)
		sc.PulseB.amplify = amplify

		sc.PulseB.frequency = int(frequency)
		sc.PulseB.dutyCycle = int(dutyCycle)
	}

	return val
}

// onNR30Write is called when the Sound Mode 3 On/Off register is written to.
func (sc *SoundController) onNR30Write(addr uint16, val uint8) uint8 {
	// Bits 6-0 are unused and always 1
	return val | 0x7F
}

// onNR32Write is called when the Sound Mode 3 Select Output Level register is
// written to.
func (sc *SoundController) onNR32Write(addr uint16, val uint8) uint8 {
	// Bits 7 and bits 4-0 are unused and always 1
	return val | 0x9F
}

// onNR34Write is called when the NR34 memory register is written to. When a 1
// is written to bit 7 of this register, the wave voice is restarted with
// the configuration found in this register and others.
func (sc *SoundController) onNR34Write(addr uint16, val uint8) uint8 {
	if val&0x80 == 0x80 {
		sc.Wave.On = true

		// Wave volume is one bit in size
		sc.Wave.volume = int(sc.state.mmu.at(nr30Addr) & 0x80 >> 7)

		// Load duration information
		duration := sc.state.mmu.at(nr31Addr)
		sc.Wave.duration = 256 - int(duration)
		sc.Wave.useDuration = val&0x40 == 0x40

		frequency := uint16(sc.state.mmu.at(nr33Addr))
		frequency |= uint16(val&0x7) << 8

		rightShiftCode := (sc.state.mmu.at(nr32Addr) & 0x60) >> 5

		sc.Wave.rightShiftCode = int(rightShiftCode)
		sc.Wave.frequency = int(frequency)
		sc.Wave.pattern = sc.readWaveTable()
	}

	return val
}

// onNR41Write is called when the Sound Mode 4 Sound Length register is written
// to.
func (sc *SoundController) onNR41Write(addr uint16, val uint8) uint8 {
	// Bits 7 and 6 are unused and always 1
	return val | 0xC0
}

// onNR44Write is called when the NR44 memory register is written to. When a 1
// is written to bit 7 of this register, the noise voice is restarted with
// the configuration found in this register and others.
func (sc *SoundController) onNR44Write(addr uint16, val uint8) uint8 {
	if val&0x80 == 0x80 {
		sc.Noise.On = true

		duration := sc.state.mmu.at(nr41Addr) & 0x3F
		sc.Noise.duration = 64 - int(duration)
		sc.Noise.useDuration = val&0x40 == 0x40

		nr42 := sc.state.mmu.at(nr42Addr)
		volume := (nr42 & 0xF0) >> 4
		amplify := nr42&0x8 == 0x8
		volumePeriod := nr42 & 0x7

		nr43 := sc.state.mmu.at(nr43Addr)
		shiftClockFrequency := (nr43 & 0xF0) >> 4
		dividingRatio := nr43 & 0x7

		var widthMode LFSRWidthMode
		if nr43&0x8 == 0x8 {
			widthMode = WidthMode7Bit
		} else {
			widthMode = WidthMode15Bit
		}

		sc.Noise.volume = int(volume)
		sc.Noise.volumePeriod = int(volumePeriod)
		sc.Noise.amplify = amplify

		sc.Noise.shiftClockFrequency = int(shiftClockFrequency)
		sc.Noise.dividingRatio = int(dividingRatio)
		sc.Noise.widthMode = widthMode
	}

	// Bits 5-0 are unused and always 1
	return val | 0x3F
}

// onNR50Write is called when the Cartridge Channel Control and Volume Register
// is written to. This register controls left and right channel audio volume.
func (sc *SoundController) onNR50Write(addr uint16, val uint8) uint8 {
	sc.leftVolume = int((val & 0x70) >> 4)
	sc.rightVolume = int(val & 0x07)

	return val
}

// onNR51Write is called when the Selection of Sound Output Terminal register
// is written to. This register enables or disables each voice on either the
// right or the left audio channel. This allows for stereo sound.
func (sc *SoundController) onNR51Write(addr uint16, val uint8) uint8 {
	sc.Noise.LeftEnabled = val&0x80 == 0x80
	sc.Wave.LeftEnabled = val&0x40 == 0x40
	sc.PulseB.LeftEnabled = val&0x20 == 0x20
	sc.PulseA.LeftEnabled = val&0x10 == 0x10
	sc.Noise.RightEnabled = val&0x08 == 0x08
	sc.Wave.RightEnabled = val&0x04 == 0x04
	sc.PulseB.RightEnabled = val&0x02 == 0x02
	sc.PulseA.RightEnabled = val&0x01 == 0x01

	return val
}

// onNR52Write is called when the Sound On/Off register is written to. On
// write, it can enable or disable the sound.
func (sc *SoundController) onNR52Write(addr uint16, val uint8) uint8 {
	sc.Enabled = val&0x80 == 0x80

	// TODO(velovix): Zero out all registers except length and stop receiving
	// writes
	// TODO(velovix): Make the "on" values for channels available here

	// Bits 6-4 are unused and always 1
	return val | 0x70
}

func (sc *SoundController) readWaveTable() []uint8 {
	wavePattern := make([]uint8, 0, 32)

	for addr := uint16(wavePatternRAMStart); addr < wavePatternRAMEnd; addr++ {
		val := sc.state.mmu.at(addr)
		lower, upper := split(val)
		wavePattern = append(wavePattern, upper, lower)
	}

	return wavePattern
}
