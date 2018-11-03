// Package main is a Game Boy emulator.
package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
)

// printInstructions controls whether or not instructions are printed. This is
// useful for debugging but slows emulation to a crawl.
const printInstructions = true

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
		os.Exit(1)
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
	var mmu mmu
	switch header.cartridgeType {
	case romOnlyCartridgeType:
		mmu = newROMOnly(cartridgeData)
	case mbc1CartridgeType:
		mmu = newMBC1(cartridgeData, header)
	default:
		fmt.Println("Error: Unknown cartridge type", header.cartridgeType)
		os.Exit(1)
	}

	env := newEnvironment(mmu)

	// Start the timers
	timers := newTimers(env)

	// Start the display controller
	vc, err := newVideoController(env, timers)
	if err != nil {
		fmt.Println("Error while creating video controller:", err)
		os.Exit(1)
	}
	defer vc.destroy()

	// Memory dump on SIGINT
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go dumpOnSigint(env, sigint)

	// Start running
	err = startMainLoop(env, &vc, timers)
	if err != nil {
		fmt.Println("Error while running ROM:", err)
		os.Exit(1)
	}
}

// dumpOnSigint dumps the current memory to "memory.dump" when something is
// received on the notifier channel.
func dumpOnSigint(env *environment, notifier chan os.Signal) {
	for range notifier {
		fmt.Println("\n\nSIGINT captured, dumping memory")

		dump, err := os.Create("memory.dump")
		if err != nil {
			fmt.Println("Error while creating dump file: %v", err)
			os.Exit(1)
		}
		defer dump.Close()

		dump.Write(env.mmu.dump())
		fmt.Println("Done dumping (tee hee!)")

		break
	}

	os.Exit(1)

}
