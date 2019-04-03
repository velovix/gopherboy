package main

import (
	"syscall/js"
)

type message struct {
	kind string
	data js.Value
}

type eventHandler struct {
	subscribers []chan message
}

func (eh *eventHandler) onMessage(this js.Value, args []js.Value) interface{} {
	event := args[0]

	event.Call("preventDefault")

	msg := message{
		kind: event.Get("data").Index(0).String(),
		data: event.Get("data").Index(1),
	}

	for _, subscriber := range eh.subscribers {
		subscriber <- msg
	}

	return nil
}
