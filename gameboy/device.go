package gameboy

import (
	"fmt"
	"time"

	"golang.org/x/xerrors"
)

// Device represents the Game Boy hardware.
type Device struct {
	state            *State
	header           romHeader
	debugger         *debugger
	timers           *timers
	videoController  *videoController
	joypad           *joypad
	serial           *serial
	interruptManager *interruptManager
	SoundController  *SoundController
	opcodeMapper     *opcodeMapper

	saveGames SaveGameDriver
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
	saveGames SaveGameDriver,
	dbConfig DebugConfiguration) (*Device, error) {

	var device Device

	device.saveGames = saveGames

	device.header = loadROMHeader(cartridgeData)
	fmt.Printf("%+v\n", device.header)

	batteryBacked := false

	// Create a memory bank controller for this ROM
	var mbc mbc
	switch device.header.cartridgeType {
	case 0x00:
		// ROM ONLY
		mbc = newROMOnlyMBC(device.header, cartridgeData)
	case 0x01:
		// MBC1
		mbc = newMBC1(device.header, cartridgeData)
	case 0x02:
		// MBC1+RAM
		mbc = newMBC1(device.header, cartridgeData)
	case 0x03:
		// MBC1+RAM+BATTERY
		mbc = newMBC1(device.header, cartridgeData)
		batteryBacked = true
		// case 0x05:
		// MBC2
		// TODO(velovix): Support this
		// case 0x06:
		// MBC2+BATTERY
		// TODO(velovix): Support this
		// case 0x08:
		// ROM+RAM
		// TODO(velovix): Support this
		// case 0x09:
		// ROM+RAM+BATTERY
		// TODO(velovix): Support this
		// case 0x0B:
		// MMM01
		// TODO(velovix): Support this
		// case 0x0C:
		// MMM01+RAM
		// TODO(velovix): Support this
		// case 0x0D:
		// MMM01+RAM+BATTERY
		// TODO(velovix): Support this
		// case 0x0F:
		// MBC3+RTC+BATTERY
		// TODO(velovix): Support this
		// batteryBacked = true
		// case 0x10:
		// MBC3+RTC+RAM+BATTERY
		// TODO(velovix): Support this
		batteryBacked = true
	case 0x11:
		// MBC3
		mbc = newMBC3(device.header, cartridgeData, false)
	case 0x12:
		// MBC3+RAM
		mbc = newMBC3(device.header, cartridgeData, false)
	case 0x13:
		// MBC3+RAM+BATTERY
		mbc = newMBC3(device.header, cartridgeData, false)
		batteryBacked = true
	case 0x19:
		// MBC5
		mbc = newMBC5(device.header, cartridgeData, false)
	case 0x1A:
		// MBC5+RAM
		mbc = newMBC5(device.header, cartridgeData, false)
	case 0x1B:
		// MBC5+RAM+BATTERY
		mbc = newMBC5(device.header, cartridgeData, false)
		batteryBacked = true
	case 0x1C:
		// MBC5+RUMBLE
		mbc = newMBC5(device.header, cartridgeData, true)
	case 0x1D:
		// MBC5+RUMBLE+RAM
		mbc = newMBC5(device.header, cartridgeData, true)
	case 0x1E:
		// MBC5+RUMBLE+RAM+BATTERY
		mbc = newMBC5(device.header, cartridgeData, true)
		batteryBacked = true
	// case 0x20:
	// MBC6
	// TODO(velovix): Support this
	// case 0x22:
	// MBC7+SENSOR+RUMBLE+RAM+BATTERY
	// TODO(velovix): Support this
	// case 0xFC:
	// POCKET CAMERA
	// TODO(velovix): Support this
	// case 0xFD:
	// BANDAI TAMA5
	// TODO(velovix): Support this
	// case 0xFE:
	// HuC3
	// TODO(velovix): Support this
	// case 0xFF:
	// HuC1+RAM+BATTERY
	// TODO(velovix): Support this
	// batteryBacked = true
	default:
		return nil, xerrors.Errorf("unknown cartridge type %#x", device.header.cartridgeType)
	}

	// Load up a save game if we're using a battery backed cartridge
	if batteryBacked {
		batteryMBC := mbc.(batteryBackedMBC)
		hasSave, err := device.saveGames.Has(device.header.title)
		if err != nil {
			return nil, xerrors.Errorf("checking for saves: %w", err)
		}
		if hasSave {
			fmt.Println("Loading battery-backed game save...")
			data, err := device.saveGames.Load(device.header.title)
			if err != nil {
				return nil, xerrors.Errorf("loading game save: %w", err)
			}
			batteryMBC.loadBatteryBackedRAM(data)
		}
	}

	mmu := newMMU(bootROM, cartridgeData, mbc)
	device.state = NewState(mmu)

	if dbConfig.Debugging {
		device.debugger = &debugger{state: device.state}

		device.debugger.breakOnPC = dbConfig.BreakOnPC
		device.debugger.breakOnOpcode = dbConfig.BreakOnOpcode
		device.debugger.breakOnAddrRead = dbConfig.BreakOnAddrRead
		device.debugger.breakOnAddrWrite = dbConfig.BreakOnAddrWrite

		device.state.mmu.db = device.debugger
	}

	device.timers = newTimers(device.state)
	mmu.timers = device.timers

	device.videoController = newVideoController(
		device.state, video)
	mmu.videoController = device.videoController

	device.joypad = newJoypad(device.state, input)

	device.serial = newSerial(device.state)

	device.interruptManager = newInterruptManager(device.state, device.timers)
	device.joypad.interruptManager = device.interruptManager
	device.videoController.interruptManager = device.interruptManager
	device.timers.interruptManager = device.interruptManager
	mmu.interruptManager = device.interruptManager

	device.SoundController = newSoundController(device.state)

	device.opcodeMapper = newOpcodeMapper(device.state)

	return &device, nil
}

// BenchmarkComponents prints out performance information on each component of
// the device. Drivers are temporarily mocked out. This will leave the device
// in a strange state that will likely not play nicely with games.
//
// Once I'm more confident of the performance of this emulator, this could
// likely be removed.
func (device *Device) BenchmarkComponents() {
	secondCycles := 10

	start := time.Now()
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			device.timers.tick()
		}
	}
	fmt.Println("Timer performance:", float64(secondCycles)/time.Since(start).Seconds())

	start = time.Now()
	oldVideoDriver := device.videoController.driver
	device.videoController.driver = &noopVideoDriver{}
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			device.videoController.tick()
		}
	}
	fmt.Println("Video controller performance:", float64(secondCycles)/time.Since(start).Seconds())
	device.videoController.driver = oldVideoDriver

	start = time.Now()
	oldInputDriver := device.joypad.driver
	device.joypad.driver = &noopInputDriver{}
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			device.joypad.tick()
		}
	}
	device.joypad.driver = oldInputDriver
	fmt.Println("Joypad performance:", float64(secondCycles)/time.Since(start).Seconds())

	start = time.Now()
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			device.interruptManager.check()
		}
	}
	fmt.Println("Interrupt manager performance:", float64(secondCycles)/time.Since(start).Seconds())

	start = time.Now()
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			device.SoundController.tick()
		}
	}
	fmt.Println("Sound controller performance:", float64(secondCycles)/time.Since(start).Seconds())

	start = time.Now()
	for i := 0; i < secondCycles; i++ {
		for j := 0; j < cpuClockRate; j++ {
			//device.opcodeMapper.run(0x00)
		}
	}
	fmt.Println("Opcode mapper performance:", float64(secondCycles)/time.Since(start).Seconds())
}

// Start starts the main processing loop of the Gameboy.
func (device *Device) Start(onExit chan bool) error {
	var currentInstruction instruction
	var err error

	for {
		// Check periodically (but not every tick for performance reasons) if
		// the main loop should be exited
		if device.timers.cpuClock == 0 {
			select {
			case <-onExit:
				// Save the game, if necessary
				if mbc, ok := device.state.mmu.mbc.(batteryBackedMBC); ok {
					fmt.Println("Saving battery-backed game state...")
					data := mbc.dumpBatteryBackedRAM()
					err := device.saveGames.Save(device.header.title, data)
					if err != nil {
						return xerrors.Errorf("saving game: %w", err)
					}
				}
				return nil
			default:
			}
		}

		device.joypad.tick()
		if device.state.stopped {
			// We're in stop mode, don't do anything
			time.Sleep(time.Millisecond)
			continue
		}

		device.timers.tick()

		if device.state.halted {
			// The device is halted. Process no new instructions, but check for
			// interrupts.
			device.interruptManager.check()
		} else if !device.state.halted {
			// Notify the debugger that we're at this PC value
			/*if device.debugger != nil {
				device.debugger.pcHook(device.state.regPC.get())
			}*/

			if currentInstruction == nil {
				// Process interrupts before fetching a new instruction. Note
				// that this means interrupt processing does not happen while
				// an instruction is being executed
				// TODO(velovix): Is this the right behavior?
				device.interruptManager.check()

				// Fetch a new operation
				opcode := device.state.incrementPC()

				if device.debugger != nil {
					device.debugger.opcodeHook(opcode)
				}

				currentInstruction, err = device.opcodeMapper.getInstruction(opcode)
				if err != nil {
					return err
				}
			}

			// Get the next step in the instruction
			currentInstruction = currentInstruction(device.state)
		}

		device.state.mmu.tick()
		device.videoController.tick()
		device.SoundController.tick()

		// Process any delayed requests to toggle the master interrupt switch.
		// These are created by the EI and DI instructions.
		if device.state.enableInterruptsTimer > 0 {
			device.state.enableInterruptsTimer--
			if device.state.enableInterruptsTimer == 0 {
				device.state.interruptsEnabled = true
			}
		}
	}
}

const (
	// interruptDispatchMCycles is the number of M-Cycles consumed while an
	// interrupt is being dispatched.
	interruptDispatchMCycles = 5
)
