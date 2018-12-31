package gameboy

import (
	"fmt"
	"time"
)

type Device struct {
	state           *State
	header          romHeader
	debugger        *debugger
	timers          *timers
	videoController *videoController
	joypad          *joypad
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

	// Create a memory bank controller for this ROM
	var mbc mbc
	switch device.header.cartridgeType {
	case 0x00:
		// ROM ONLY
		mbc = newROMOnlyMBC(cartridgeData)
	case 0x01:
		// MBC1
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

		device.joypad.tick()
		if device.state.stopped {
			// We're in stop mode, don't do anything
			time.Sleep(time.Millisecond)
			continue
		}

		if device.state.waitingForInterrupts {
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

		// Check if any interrupts need to be processed
		if device.state.interruptsEnabled && device.state.mmu.at(ifAddr) != 0 {
			var target uint16

			interruptEnable := device.state.mmu.at(ieAddr)
			interruptFlag := device.state.mmu.at(ifAddr)

			// Check each bit of the interrupt flag to see if an interrupt
			// happened, and each bit of the interrupt enable flag to check if
			// we should process it. Then, reset the interrupt flag.
			if interruptEnable&interruptFlag&0x01 == 0x01 {
				// VBlank interrupt
				target = vblankInterruptTarget
				interruptFlag &= ^uint8(0x01)
			} else if interruptEnable&interruptFlag&0x02 == 0x02 {
				// LCDC interrupt
				target = lcdcInterruptTarget
				interruptFlag &= ^uint8(0x02)
			} else if interruptEnable&interruptFlag&0x04 == 0x04 {
				// TIMA overflow interrupt
				target = timaOverflowInterruptTarget
				interruptFlag &= ^uint8(0x04)
			} else if interruptEnable&interruptFlag&0x08 == 0x08 {
				// Serial interrupt
				target = serialInterruptTarget
				interruptFlag &= ^uint8(0x08)
			} else if interruptEnable&interruptFlag&0x10 == 0x10 {
				// P10-P13 interrupt
				target = p1Thru4InterruptTarget
				interruptFlag &= ^uint8(0x10)
			}

			device.state.mmu.setNoNotify(ifAddr, interruptFlag)

			if target != 0 {
				// Disable all other interrupts
				device.state.interruptsEnabled = false
				device.state.waitingForInterrupts = false
				// Push the current program counter to the stack for later use
				device.state.pushToStack16(device.state.regs16[regPC].get())
				// Jump to the target
				device.state.regs16[regPC].set(target)
			}
		}

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
