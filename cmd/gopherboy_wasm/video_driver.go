package main

import (
	"syscall/js"
)

// videoDriver provides a video driver interface with WebGL as its back end.
// This can be used inside a WebGL-capable browser as a WebAssembly
// application.
type videoDriver struct {
}

func newVideoDriver(scaleFactor float64) (*videoDriver, error) {
	var vd videoDriver
	return &vd, nil
}

func (vd *videoDriver) Clear() {
}

func (vd *videoDriver) Render(frameData []uint8) error {
	jsFrameData := js.TypedArrayOf(frameData)
	defer jsFrameData.Release()

	jsClamped := js.Global().Get("Uint8ClampedArray").New(jsFrameData)

	js.Global().Call("postMessage",
		[]interface{}{
			"NewFrame",
			jsClamped,
		},
	)
	return nil
}

func (vd *videoDriver) Close() {

}
