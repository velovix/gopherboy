package main

import (
	"fmt"

	"github.com/velovix/gopherboy/gameboy"
)

type inputDriver struct {
	buttonStates map[gameboy.Button]bool

	messages chan message
}

func newInputDriver() *inputDriver {
	return &inputDriver{
		buttonStates: make(map[gameboy.Button]bool),
		messages:     make(chan message, 20),
	}
}

func (driver *inputDriver) State(btn gameboy.Button) bool {
	return driver.buttonStates[btn]
}

func (driver *inputDriver) Update() bool {
	newButtonPressed := false

	select {
	case msg := <-driver.messages:
		switch msg.kind {
		case "ButtonPressed":
			newButtonPressed = true
			driver.buttonStates[gameboy.Button(msg.data.Int())] = true
		case "ButtonReleased":
			driver.buttonStates[gameboy.Button(msg.data.Int())] = false
		default:
			fmt.Println("emulator: Ignoring message", msg)
		}
	default:
	}

	return newButtonPressed
}
