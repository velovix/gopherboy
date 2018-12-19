package main

import (
	"fmt"
	"time"

	"github.com/velovix/gopherboy/gameboy"
)

func main() {
	//dropZone := js.Global().Get("document").Call("getElementById", "rom-drop-zone")

	onCartridgeData := make(chan []byte)

	/*dropCallback := js.NewCallback(
		func(args []js.Value) {
			items := event.Get("dataTransfer").Get("items")
			fmt.Println("All items", items)
			fmt.Println(items.Get("length"))

			for i := 0; i < items.Length(); i++ {
				if items.Index(i).Get("kind").String() == "file" {
					fmt.Println("got file!")
				} else {
					fmt.Println("Ignoring non-file")
				}
			}
			onCartridgeData <- []byte{1, 2, 3}
		})

	dropZone.Call("addEventListener", "drop", dropCallback)*/

	time.Sleep(time.Second * 5)

	cartridgeData := <-onCartridgeData

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
