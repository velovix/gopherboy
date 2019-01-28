package main

import (
	"fmt"
	"syscall/js"

	"github.com/velovix/gopherboy/gameboy"
)

func main() {
	fmt.Println("emulator: Waiting for cartridge data")

	// Set up the event handler
	eventHandler := eventHandler{}
	js.Global().Set("onmessage", js.NewEventCallback(0, eventHandler.onMessage))

	var cartridgeData []byte

	// Wait for cartridge data
	cartridgeDataEvents := make(chan message)
	eventHandler.subscribers = append(eventHandler.subscribers, cartridgeDataEvents)
	for cartridgeData == nil {
		msg := <-cartridgeDataEvents

		switch msg.kind {
		case "CartridgeData":
			cartridgeData = make([]uint8, msg.data.Length())
			for i := 0; i < msg.data.Length(); i++ {
				cartridgeData[i] = uint8(msg.data.Index(i).Int())
			}

			fmt.Println("emulator: Cartridge data message received")
		default:
			fmt.Println("emulator: Ignoring message", msg)
		}
	}
	eventHandler.subscribers = []chan message{}

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
	eventHandler.subscribers = append(eventHandler.subscribers, input.messages)

	device, err := gameboy.NewDevice(gameboy.BootROM(), cartridgeData, video, input, &mockSaveGameDriver{}, gameboy.DebugConfiguration{})
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
