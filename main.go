package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Usage: gopherboy [OPTIONS] rom_file")
		os.Exit(1)
	}

	romFile, err := os.Open(flag.Args()[0])
	if err != nil {
		fmt.Println("Error while opening ROM file:", err)
		os.Exit(1)
	}
	err = loadROM(romFile)
	if err != nil {
		fmt.Println("Error while reading ROM:", err)
		os.Exit(1)
	}
}
