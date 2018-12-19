// +build wasm

package main

type joypad struct{}

func newJoypad(state *State) *joypad {
	return &joypad{}
}

func (j *joypad) tick() {
}
