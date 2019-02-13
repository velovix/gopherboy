package main

// typedef unsigned char Uint8;
// void SoundCallback(void *userdata, Uint8 *stream, int len);
import "C"
import (
	"log"
	"math"
	"reflect"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/velovix/gopherboy/gameboy"
)

const (
	toneHz   = 440
	sampleHz = 128
	totalHz  = 44100
)

var myDevice *gameboy.Device
var samples = make(chan float64, totalHz*3)

var (
	pulseAPhase = 0.0
	pulseBPhase = 0.0
	wavePhase   = 0.0
	shiftCount  = 0.0
	lfsr        = uint16(0x38C2)
)

const tau = math.Pi * 2.0

func square(phase, dutyCycle float64) float64 {
	phase -= tau * math.Floor(phase/tau)

	if phase < tau*dutyCycle {
		return 1.0
	}
	return 0.0
}

func wave(phase float64, pattern [32]float64) float64 {
	wavePos := math.Floor((math.Mod(phase, tau) / tau) * 32)

	return pattern[int(wavePos)]
}

//export SoundCallback
func SoundCallback(userdata unsafe.Pointer, stream *C.Uint8, length C.int) {
	n := int(length)
	sliceHeader := reflect.SliceHeader{Data: uintptr(unsafe.Pointer(stream)), Len: n, Cap: n}
	buffer := *(*[]C.Uint8)(unsafe.Pointer(&sliceHeader))

	pulseAPhaseDelta := tau * float64(myDevice.SoundController.PulseA.Frequency()) / totalHz
	pulseBPhaseDelta := tau * myDevice.SoundController.PulseB.Frequency() / totalHz
	wavePhaseDelta := tau * myDevice.SoundController.Wave.Frequency() / totalHz

	if !myDevice.SoundController.Enabled {
		// Fill the buffer with zeros
		for i := 0; i < n; i++ {
			buffer[i] = 0
		}
		return
	}

	for i := 0; i < n; i += 2 {
		var sampleLeft, sampleRight float64

		if myDevice.SoundController.PulseA.On {
			pulseAPhase += pulseAPhaseDelta

			sample := square(
				pulseAPhase,
				myDevice.SoundController.PulseA.DutyCycle())
			sample *= myDevice.SoundController.PulseA.Volume()

			if myDevice.SoundController.PulseA.LeftEnabled {
				sampleLeft += sample
			}
			if myDevice.SoundController.PulseA.RightEnabled {
				sampleRight += sample
			}
		}

		if myDevice.SoundController.PulseB.On {
			pulseBPhase += pulseBPhaseDelta

			sample := square(
				pulseBPhase,
				myDevice.SoundController.PulseB.DutyCycle())
			sample *= myDevice.SoundController.PulseB.Volume()

			if myDevice.SoundController.PulseB.LeftEnabled {
				sampleLeft += sample
			}
			if myDevice.SoundController.PulseB.RightEnabled {
				sampleRight += sample
			}
		}

		if myDevice.SoundController.Wave.On {
			wavePhase += wavePhaseDelta

			sample := wave(
				wavePhase,
				myDevice.SoundController.Wave.Pattern())
			sample *= myDevice.SoundController.Wave.Volume()

			if myDevice.SoundController.Wave.LeftEnabled {
				sampleLeft += sample
			}
			if myDevice.SoundController.Wave.RightEnabled {
				sampleRight += sample
			}
		}

		if myDevice.SoundController.Noise.On {
			shiftCount += myDevice.SoundController.Noise.ShiftFrequency() / totalHz

			for shiftCount >= 1 {
				lfsr >>= 1
				xorVal := (lfsr & 0x1) ^ ((lfsr & 0x2) >> 1)
				lfsr |= xorVal << 14

				if myDevice.SoundController.Noise.WidthMode() == gameboy.WidthMode7Bit {
					lfsr |= xorVal << 6
				}
				shiftCount--
			}

			sample := float64(lfsr&0x1) * myDevice.SoundController.Noise.Volume()

			if myDevice.SoundController.Noise.LeftEnabled {
				sampleLeft += sample
			}
			if myDevice.SoundController.Noise.RightEnabled {
				sampleRight += sample
			}
		}

		sampleLeft *= myDevice.SoundController.LeftVolume()
		sampleRight *= myDevice.SoundController.RightVolume()

		buffer[i] = C.Uint8(uint8(sampleLeft * 50))
		buffer[i+1] = C.Uint8(uint8(sampleRight * 50))
	}
}

type soundDriver struct{}

func newSoundDriver(device *gameboy.Device) (*soundDriver, error) {
	myDevice = device

	spec := &sdl.AudioSpec{
		Freq:     totalHz,
		Format:   sdl.AUDIO_U8,
		Channels: 2,
		Samples:  sampleHz,
		Callback: sdl.AudioCallback(C.SoundCallback),
	}
	err := sdl.OpenAudio(spec, nil)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	sdl.PauseAudio(false)

	return &soundDriver{}, nil
}
