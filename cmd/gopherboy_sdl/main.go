package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"runtime"

	"github.com/pkg/profile"
	"github.com/velovix/gopherboy/gameboy"
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
	_ = unlimitedFPS

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

	video, err := newVideoDriver(*scaleFactor)
	if err != nil {
		fmt.Println("Error: While initializing video driver:", err)
		os.Exit(1)
	}
	input := newInputDriver()
	if err != nil {
		fmt.Println("Error: While initializing input driver:", err)
		os.Exit(1)
	}

	var dbConfig gameboy.DebugConfiguration

	if *breakOnOpcode != -1 || *breakOnAddrRead != -1 || *breakOnAddrWrite != -1 {
		dbConfig.Debugging = true
		if *breakOnOpcode != -1 {
			val := uint8(*breakOnOpcode)
			dbConfig.BreakOnOpcode = &val
		}
		if *breakOnAddrRead != -1 {
			val := uint16(*breakOnAddrRead)
			dbConfig.BreakOnAddrRead = &val
		}
		if *breakOnAddrWrite != -1 {
			val := uint16(*breakOnAddrWrite)
			dbConfig.BreakOnAddrWrite = &val
		}
	}

	device, err := gameboy.NewDevice(cartridgeData, video, input, dbConfig)
	if err != nil {
		fmt.Println("Error: While initializing Game Boy:", err)
		os.Exit(1)
	}

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

	err = device.Start(onExit)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}

	fmt.Println("Buh-bye!")
}