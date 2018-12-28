package main

import (
	"fmt"
	"syscall/js"

	"github.com/velovix/gopherboy/gameboy"
)

func main() {
	fmt.Println("emulator: Waiting for cartridge data")

	cartridgeDataChan := make(chan []byte)

	onMessage := js.NewEventCallback(0, func(event js.Value) {
		fmt.Println("emulator: Cartridge data message received")

		jsData := event.Get("data")

		data := make([]uint8, jsData.Length())
		for i := 0; i < jsData.Length(); i++ {
			data[i] = uint8(jsData.Index(i).Int())
		}
		cartridgeDataChan <- data
	})

	js.Global().Set("onmessage", onMessage)

	cartridgeData := <-cartridgeDataChan

	fmt.Println("emulator: Main thread received cartridge data")

	video, err := newVideoDriver(2)
	if err != nil {
		fmt.Println("Error: While initializing video driver:", err)
		return
	}
	input := newInputDriver()
	if err != nil {
		fmt.Println("Error: While initializing input driver:", err)
		return
	}

	device, err := gameboy.NewDevice(cartridgeData, video, input, gameboy.DebugConfiguration{})
	if err != nil {
		fmt.Println("Error: While initializing Game Boy:", err)
		return
	}

	onExit := make(chan bool)

	err = device.Start(onExit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
