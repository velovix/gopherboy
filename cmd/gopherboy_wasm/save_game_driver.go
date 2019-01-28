package main

import "fmt"

type mockSaveGameDriver struct{}

func (driver *mockSaveGameDriver) Save(name string, data []uint8) error {
	fmt.Println("Save games are not yet supported in WASM")
	return nil
}

func (driver *mockSaveGameDriver) Load(name string) ([]uint8, error) {
	panic("loading games is not yet supported in WASM")
}

func (driver *mockSaveGameDriver) Has(name string) (bool, error) {
	return true, nil
}
