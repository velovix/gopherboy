package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
		os.Exit(1)
	}

	env := newEnvironment()

	// Load up the ROM
	romFile, err := os.Open(flag.Args()[0])
	if err != nil {
		fmt.Println("Error while opening ROM file:", err)
		os.Exit(1)
	}
	err = loadROM(env, romFile)
	if err != nil {
		fmt.Println("Error while reading ROM:", err)
		os.Exit(1)
	}

	// Start the display controller
	vc, err := newVideoController(env)
	if err != nil {
		fmt.Println("Error while creating video controller:", err)
		os.Exit(1)
	}
	defer vc.destroy()

	// Memory dump on SIGINT
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	go func() {
		for range sigint {
			fmt.Println("\n\nSIGINT captured, dumping memory")

			dump, err := os.Create("memory.dump")
			if err != nil {
				fmt.Println("while creating dump file: %v", err)
				os.Exit(1)
			}
			defer dump.Close()

			dump.Write(env.mem)
			fmt.Println("Done dumping (tee hee!)")

			break
		}

		os.Exit(1)
	}()

	// Start running
	err = startMainLoop(env, &vc)
	if err != nil {
		fmt.Println("Error while running ROM:", err)
		os.Exit(1)
	}
}
