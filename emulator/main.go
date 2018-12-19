// Package main is a Game Boy emulator.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"

	"github.com/pkg/profile"
)

var (
	// printInstructions controls whether or not instructions are printed. This is
	// useful for debugging but slows emulation to a crawl.
	printInstructions = false
)

func main() {
	// Necessary because SDL2 uses OpenGL in the back end, which is not
	// thread-safe and does not like being moved around
	runtime.LockOSThread()

	scaleFactor := flag.Float64("scale", 2,
		"The amount to scale the window by, with 1 being native resolution")
	breakOnOpcode := flag.Int("break-on-opcode", -1,
		"An opcode to break at")
	breakOnAddrRead := flag.Int("break-on-addr-read", -1,
		"A memory address to break at on read")
	breakOnAddrWrite := flag.Int("break-on-addr-write", -1,
		"A memory address to break at on write")
	enableProfiling := flag.Bool("profile", false,
		"Generates a pprof file if set")
	unlimitedFPS := flag.Bool("unlimited-fps", false,
		"If true, frame rate will not be capped. Games will run as quickly as possible.")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
		os.Exit(1)
	}

	if *scaleFactor <= 0 {
		fmt.Println("Scale factor must be higher than 0")
		os.Exit(1)
	}

	if *enableProfiling {
		fmt.Println("Profiling has been enabled")
		defer profile.Start(profile.NoShutdownHook).Stop()
	}

	// Load the ROM file
	cartridgeData, err := ioutil.ReadFile(flag.Args()[0])
	if err != nil {
		fmt.Println("Error: While reading cartridge:", err)
		os.Exit(1)
	}

	// Load the header
	header := loadROMHeader(cartridgeData)
	fmt.Println(header)

	// Create a memory bank controller for this ROM
	var mbc mbc
	switch header.cartridgeType {
	case 0x00:
		// ROM ONLY
		mbc = newROMOnlyMBC(cartridgeData)
	case 0x01:
		// MBC1
		mbc = newMBC1(header, cartridgeData)
	case 0x13:
		// MBC3+RAM+BATTERY
		mbc = newMBC3(header, cartridgeData, false)
	default:
		fmt.Printf("Error: Unknown cartridge type %#x\n", header.cartridgeType)
		os.Exit(1)
	}

	state := NewState(newMMU(cartridgeData, mbc))

	// Start the debugger
	var db *debugger
	if *breakOnOpcode != -1 || *breakOnAddrRead != -1 || *breakOnAddrWrite != -1 {
		printInstructions = true
		fmt.Println("Setting up debugger")
		db = &debugger{state: state}

		if *breakOnOpcode != -1 {
			val := uint8(*breakOnOpcode)
			db.breakOnOpcode = &val
		}
		if *breakOnAddrRead != -1 {
			val := uint16(*breakOnAddrRead)
			db.breakOnAddrRead = &val
		}
		if *breakOnAddrWrite != -1 {
			val := uint16(*breakOnAddrWrite)
			db.breakOnAddrWrite = &val
		}

		state.mmu.db = db
	}

	// Start the timers
	timers := newTimers(state)

	// Start the display controller
	vc, err := newVideoController(state, timers, *scaleFactor)
	if err != nil {
		fmt.Println("Error while creating video controller:", err)
		os.Exit(1)
	}
	defer vc.destroy()

	vc.unlimitedFPS = *unlimitedFPS

	// Start reading joypad input
	joypad := newJoypad(state)

	// Stop main loop on sigint
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	onExit := make(chan bool)
	go func() {
		for range sigint {
			onExit <- true
			break
		}
	}()

	// Start running
	err = startMainLoop(state, vc, timers, joypad, db, onExit)
	if err != nil {
		fmt.Println("Error while running ROM:", err)
		os.Exit(1)
	}

	fmt.Println("Dumping memory")

	dump, err := os.Create("memory.dump")
	if err != nil {
		fmt.Println("Error while creating dump file:", err)
		os.Exit(1)
	}
	defer dump.Close()

	dump.Write(state.mmu.dump())
	fmt.Println("Done dumping (tee hee!)")
}
