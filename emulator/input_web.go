// +build wasm

package main

type button int

const (
	buttonStart button = iota
	buttonSelect
	buttonB
	buttonA
	buttonDown
	buttonUp
	buttonLeft
	buttonRight
)

type inputDriver struct {
	buttonStates map[button]bool
}

func newInputDriver() *inputDriver {
	return &inputDriver{}
}

func (driver *inputDriver) update() bool {
	return false
}
