// +build wasm

package main

type joypad struct{}

func newJoypad(env *environment) *joypad {
	return &joypad{}
}

func (j *joypad) tick() {
}
