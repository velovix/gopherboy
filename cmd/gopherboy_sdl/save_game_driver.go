package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

type fileSaveGameDriver struct {
	directory string
}

func (driver *fileSaveGameDriver) Save(name string, data []uint8) error {
	err := ioutil.WriteFile(driver.nameToPath(name), data, 0644)
	if err != nil {
		return fmt.Errorf("saving game save: %v", err)
	}

	return nil
}

func (driver *fileSaveGameDriver) Load(name string) ([]uint8, error) {
	filename := driver.nameToPath(name)
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("loading game save %v: %v", filename, err)
	}

	return data, nil
}

func (driver *fileSaveGameDriver) Has(name string) (bool, error) {
	filename := driver.nameToPath(name)

	_, err := os.Stat(filename)
	if err == nil {
		return true, nil
	} else if os.IsNotExist(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("checking if game save %v exists under name %v: %v",
			name, filename, err)
	}
}

func (driver *fileSaveGameDriver) nameToPath(name string) string {
	return path.Join(driver.directory, name+".sav")
}
