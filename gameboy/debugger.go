package gameboy

import (
	"bufio"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
)

// TODO(velovix): Consider making this feature available again
var printInstructions = false

type debugger struct {
	state *State

	breakOnPC        *uint16
	breakOnOpcode    *uint8
	breakOnAddrRead  *uint16
	breakOnAddrWrite *uint16
	stepping         bool
}

func (db *debugger) pcHook(pc uint16) {
	if db.breakOnPC != nil && pc == *db.breakOnPC {
		fmt.Printf("PC BREAK: %#04x\n", pc)
		db.stepping = true
	}
}

func (db *debugger) opcodeHook(opcode uint8) {
	if (db.breakOnOpcode != nil && opcode == *db.breakOnOpcode) || db.stepping {
		fmt.Printf("Opcode BREAK: %#02x\n", opcode)
		db.printState()
		db.readCommand()
	}
}

func (db *debugger) memReadHook(addr uint16) {
	if db.breakOnAddrRead != nil && addr == *db.breakOnAddrRead {
		fmt.Printf("Address read BREAK: %#04x\n", addr)
		db.printState()
		db.readCommand()
	}
}

func (db *debugger) memWriteHook(addr uint16, val uint8) {
	if db.breakOnAddrWrite != nil && addr == *db.breakOnAddrWrite {
		fmt.Printf("Address write BREAK: %#04x with value %#04x\n", addr, val)
		db.printState()
		db.readCommand()
	}
}

func (db *debugger) printState() {
	fmt.Printf("  AF: %#04x\n", db.state.regAF.get())
	fmt.Printf("  BC: %#04x\n", db.state.regBC.get())
	fmt.Printf("  DE: %#04x\n", db.state.regDE.get())
	fmt.Printf("  HL: %#04x\n", db.state.regHL.get())
	fmt.Printf("  SP: %#04x\n", db.state.regSP.get())
	fmt.Printf("  PC: %#04x\n", db.state.instructionStart)
	fmt.Printf("  DV: %#02x\n", db.state.mmu.at(dividerAddr))
}

func (db *debugger) readCommand() {
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("Now what? ")
		command, _ := reader.ReadString('\n')
		if len(command) < 2 {
			fmt.Println("\nNo command received")
			continue
		}
		// Remove trailing newline
		command = command[:len(command)-1]

		if command == "c" {
			fmt.Println("Continuing")
			db.stepping = false
			break
		} else if command == "n" {
			fmt.Println("Stepping")
			db.stepping = true
			break
		} else if strings.HasPrefix(command, "m") {
			// Parse the command
			splitted := strings.Split(command, " ")
			if len(splitted) != 2 {
				fmt.Println("Format: m [address in hex]")
				continue
			}
			addr, err := strconv.ParseUint(splitted[1], 16, 16)
			if err != nil {
				fmt.Println("Error parsing address:", err)
				continue
			}

			// Get the address, temporarily turning off break on read to avoid
			// a break loop
			oldBreakOnAddrRead := db.breakOnAddrRead
			db.breakOnAddrRead = nil
			fmt.Printf("Value at address %#04x: %#x\n",
				uint16(addr),
				db.state.mmu.at(uint16(addr)))
			db.breakOnAddrRead = oldBreakOnAddrRead
		} else if command == "trace" {
			stacktrace := debug.Stack()
			fmt.Println(string(stacktrace))
		} else {
			fmt.Printf("Unknown command '%v'\n", command)
			continue
		}
	}
}
