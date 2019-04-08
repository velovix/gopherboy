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

func init() {
	// Necessary because SDL2 is not thread-safe and does not like being moved
	// around
	runtime.LockOSThread()
}

// mainThreadFuncs contains functions queued up to run on the main thread.
var mainThreadFuncs = make(chan func())

// doOnMainThread runs the given function in the main thread and returns when
// done.
func doOnMainThread(f func(), async bool) {
	done := make(chan bool, 1)
	mainThreadFuncs <- func() {
		f()
		done <- true
	}
	if !async {
		<-done
	}
}

func main() {
	bootROM := flag.String("boot-rom", "",
		"Path to a file containing the Game Boy boot ROM")
	scaleFactor := flag.Float64("scale", 2,
		"The amount to scale the window by, with 1 being native resolution")
	breakOnPC := flag.Int("break-on-pc", -1,
		"A program counter value to break at")
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
	saveGameDirectory := flag.String("save-game-dir", ".",
		"The directory to find save games in")
	benchmarkComponents := flag.Bool("benchmark-components", false,
		"If true, some performance information will be printed out about each "+
			"component, then the emulator will exit.")
	_ = unlimitedFPS

	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
		os.Exit(1)
	}

	var bootROMData []byte
	var err error
	if *bootROM == "" {
		fmt.Println("A boot ROM must be provided")
		os.Exit(1)
	}
	// Load the boot ROM
	bootROMData, err = ioutil.ReadFile(*bootROM)
	if err != nil {
		fmt.Println("Error: While reading boot ROM:", err)
		os.Exit(1)
	}

	if *scaleFactor <= 0 {
		fmt.Println("Scale factor must be higher than 0")
		os.Exit(1)
	}

	if stat, err := os.Stat(*saveGameDirectory); os.IsNotExist(err) {
		fmt.Println("The specified save game directory does not exist")
		os.Exit(1)
	} else if err != nil {
		fmt.Println("Error with save game directory:", err)
		os.Exit(1)
	} else if !stat.IsDir() {
		fmt.Println("The specified save game directory is not a directory")
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

	video, err := newVideoDriver(*scaleFactor, *unlimitedFPS)
	if err != nil {
		fmt.Println("Error: While initializing video driver:", err)
		os.Exit(1)
	}
	input := newInputDriver()
	if err != nil {
		fmt.Println("Error: While initializing input driver:", err)
		os.Exit(1)
	}
	saveGames := &fileSaveGameDriver{
		directory: *saveGameDirectory,
	}

	var dbConfig gameboy.DebugConfiguration

	if *breakOnPC != -1 || *breakOnOpcode != -1 || *breakOnAddrRead != -1 ||
		*breakOnAddrWrite != -1 {

		dbConfig.Debugging = true
		if *breakOnPC != -1 {
			val := uint16(*breakOnPC)
			dbConfig.BreakOnPC = &val
		}
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

	device, err := gameboy.NewDevice(bootROMData, cartridgeData, video, input, saveGames, dbConfig)
	if err != nil {
		fmt.Println("Error: While initializing Game Boy:", err)
		os.Exit(1)
	}

	_, err = newSoundDriver(device)
	if err != nil {
		fmt.Println("Error: While initializing sound driver:", err)
		os.Exit(1)
	}

	if *benchmarkComponents {
		device.BenchmarkComponents()
		return
	}

	// Stop main loop on sigint
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	onExitRequest := make(chan bool)
	go func() {
		for range sigint {
			onExitRequest <- true
			break
		}
	}()

	onDeviceExit := make(chan bool)

	// Start the device
	go func() {
		err = device.Start(onExitRequest)
		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(1)
		}

		onDeviceExit <- true
	}()

	running := true
	for running {
		select {
		case f := <-mainThreadFuncs:
			f()
		case <-onDeviceExit:
			running = false
		}
	}

	fmt.Println("Buh-bye!")
}
