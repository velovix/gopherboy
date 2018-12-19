package main

import "github.com/velovix/gopherboy/gameboy"

type inputDriver struct {
	buttonStates map[gameboy.Button]bool
}

func newInputDriver() *inputDriver {
	return &inputDriver{}
}

func (driver *inputDriver) State(btn gameboy.Button) bool {
	return driver.buttonStates[btn]
}

func (driver *inputDriver) Update() bool {
	return false
}
