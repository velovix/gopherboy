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
	js.Global().Set("onmessage", js.FuncOf(eventHandler.onMessage))

	var cartridgeData []byte
	var bootROMData []byte

	// Wait for data
	dataEvents := make(chan message)
	eventHandler.subscribers = append(eventHandler.subscribers, dataEvents)

	for cartridgeData == nil || bootROMData == nil {
		msg := <-dataEvents

		switch msg.kind {
		case "BootROMData":
			bootROMData = make([]uint8, msg.data.Length())
			for i := 0; i < msg.data.Length(); i++ {
				bootROMData[i] = uint8(msg.data.Index(i).Int())
			}

			fmt.Println("emulator: Boot ROM data message received")
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

	device, err := gameboy.NewDevice(bootROMData, cartridgeData, video, input, &mockSaveGameDriver{}, gameboy.DebugConfiguration{})
	if err != nil {
		fmt.Println("Error: While initializing Game Boy:", err)
		return
	}

	onExit := make(chan bool)

	onStopMessage := make(chan message)
	eventHandler.subscribers = append(eventHandler.subscribers, onStopMessage)
	go func() {
		for {
			msg := <-onStopMessage

			if msg.kind == "Stop" {
				fmt.Println("Received stop request")
				onExit <- true
				break
			}
		}
	}()

	err = device.Start(onExit)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}
