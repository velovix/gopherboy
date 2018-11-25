// Package main is a Game Boy emulator.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"

	"github.com/pkg/profile"
)

var (
	// printInstructions controls whether or not instructions are printed. This is
	// useful for debugging but slows emulation to a crawl.
	printInstructions = false
)

func main() {
	debug := flag.Bool("debug", false, "Start in debug mode")
	breakOnOpcode := flag.Int("break-on-opcode", -1, "An opcode to break at")
	breakOnAddrRead := flag.Int("break-on-addr-read", -1, "A memory address to break at on read")
	enableProfiling := flag.Bool("profile", false, "Generates a pprof file if set")

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
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

	env := newEnvironment(newMMU(cartridgeData, mbc))

	// Start the debugger
	var db *debugger
	if *debug {
		printInstructions = true
		fmt.Println("Setting up debugger")
		db = &debugger{env: env}

		if *breakOnOpcode != -1 {
			val := uint8(*breakOnOpcode)
			db.breakOnOpcode = &val
		}
		if *breakOnAddrRead != -1 {
			val := uint16(*breakOnAddrRead)
			db.breakOnAddrRead = &val
		}

		env.mmu.db = db
	}

	// Start the timers
	timers := newTimers(env)

	// Start the display controller
	vc, err := newVideoController(env, timers)
	if err != nil {
		fmt.Println("Error while creating video controller:", err)
		os.Exit(1)
	}
	defer vc.destroy()

	// Start reading joypad input
	joypad := newJoypad(env)

	// Memory dump on SIGINT
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	onStop := make(chan bool)
	go func() {
		for range sigint {
			onStop <- true
			break
		}
	}()

	// Start running
	err = startMainLoop(env, vc, timers, joypad, db, onStop)
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

	dump.Write(env.mmu.dump())
	fmt.Println("Done dumping (tee hee!)")
}

// dumpOnSigint dumps the current memory to "memory.dump" when something is
// received on the notifier channel.
func dumpOnSigint(env *environment, notifier chan os.Signal) {
	for range notifier {

		break
	}

	os.Exit(1)

}
