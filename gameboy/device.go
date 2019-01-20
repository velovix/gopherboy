package gameboy

import (
	"fmt"
	"time"
)

type Device struct {
	state            *State
	header           romHeader
	debugger         *debugger
	timers           *timers
	videoController  *videoController
	joypad           *joypad
	serial           *serial
	interruptManager *interruptManager
	soundController  *soundController
}

type DebugConfiguration struct {
	Debugging bool

	BreakOnPC        *uint16
	BreakOnOpcode    *uint8
	BreakOnAddrRead  *uint16
	BreakOnAddrWrite *uint16
}

func NewDevice(
	bootROM []byte,
	cartridgeData []byte,
	video VideoDriver,
	input InputDriver,
	dbConfig DebugConfiguration) (*Device, error) {

	var device Device

	device.header = loadROMHeader(cartridgeData)
	fmt.Printf("%+v\n", device.header)

	// Create a memory bank controller for this ROM
	var mbc mbc
	switch device.header.cartridgeType {
	case 0x00:
		// ROM ONLY
		mbc = newROMOnlyMBC(device.header, cartridgeData)
	case 0x01:
		// MBC1
		mbc = newMBC1(device.header, cartridgeData)
	case 0x03:
		// MBC1+RAM+BATTERY
		// TODO(velovix): Add battery support
		mbc = newMBC1(device.header, cartridgeData)
	case 0x13:
		// MBC3+RAM+BATTERY
		mbc = newMBC3(device.header, cartridgeData, false)
	default:
		return nil, fmt.Errorf("unknown cartridge type %#x", device.header.cartridgeType)
	}

	device.state = NewState(newMMU(bootROM, cartridgeData, mbc))

	if dbConfig.Debugging {
		device.debugger = &debugger{state: device.state}

		device.debugger.breakOnPC = dbConfig.BreakOnPC
		device.debugger.breakOnOpcode = dbConfig.BreakOnOpcode
		device.debugger.breakOnAddrRead = dbConfig.BreakOnAddrRead
		device.debugger.breakOnAddrWrite = dbConfig.BreakOnAddrWrite

		device.state.mmu.db = device.debugger
	}

	device.timers = newTimers(device.state)

	device.videoController = newVideoController(
		device.state, device.timers, video)
	device.videoController.unlimitedFPS = false

	device.joypad = newJoypad(device.state, input)

	device.serial = newSerial(device.state)

	device.interruptManager = newInterruptManager(device.state, device.timers)

	device.soundController = newSoundController(device.state)

	return &device, nil
}

// Start starts the main processing loop of the Gameboy.
func (device *Device) Start(onExit chan bool) error {
	var opTime int

	for {
		var err error

		// Check if the main loop should be exited
		select {
		case <-onExit:
			return nil
		default:
		}

		device.interruptManager.check()

		device.joypad.tick()
		if device.state.stopped {
			// We're in stop mode, don't do anything
			time.Sleep(time.Millisecond)
			continue
		}

		if device.state.halted {
			// Spin our wheels running NOPs until an interrupt happens
			opTime = 4
		} else {
			// Notify the debugger that we're at this PC value
			if device.debugger != nil {
				device.debugger.pcHook(device.state.regs16[regPC].get())
			}

			// Fetch and run an operation
			opcode := device.state.incrementPC()

			if device.debugger != nil {
				device.debugger.opcodeHook(opcode)
			}

			opTime, err = runOpcode(device.state, opcode)
			if err != nil {
				return err
			}
			device.state.instructionDone()
		}

		device.timers.tick(opTime)
		device.state.mmu.tick(opTime)
		device.videoController.tick(opTime)

		// Process any delayed requests to toggle the master interrupt switch.
		// These are created by the EI and DI instructions.
		if device.state.enableInterruptsTimer > 0 {
			device.state.enableInterruptsTimer--
			if device.state.enableInterruptsTimer == 0 {
				device.state.interruptsEnabled = true
			}
		}
		if device.state.disableInterruptsTimer > 0 {
			device.state.disableInterruptsTimer--
			if device.state.disableInterruptsTimer == 0 {
				device.state.interruptsEnabled = false
			}
		}

	}
}

const (
	// interruptDispatchCycles is the number of CPU clock cycles consumed while an
	// interrupt is being dispatched.
	interruptDispatchCycles = 20
	// unhaltCycles is the number of CPU clock cycles consumed while taking the CPU
	// out of halt mode.
	unhaltCycles = 4
)
