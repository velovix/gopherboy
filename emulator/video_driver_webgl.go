// +build wasm

package main

import (
	"errors"
	"syscall/js"
)

// videoDriver provides a video driver interface with WebGL as its back end.
// This can be used inside a WebGL-capable browser as a WebAssembly
// application.
type videoDriver struct {
	// gl is the WebGL context of the target canvas.
	gl js.Value
}

func newVideoDriver(scaleFactor float64) (*videoDriver, error) {
	var vd videoDriver

	doc := js.Global().Get("document")
	canvas := doc.Call("getElementById", "emulator-canvas")
	vd.gl = canvas.Call("getContext", "webgl")
	if vd.gl == js.Undefined() {
		return nil, errors.New("browser does not support WebGL")
	}

	vd.gl.Call("clearColor", 1.0, 0.0, 0.0, 1.0)
	vd.gl.Call("enable", vd.gl.Get("DEPTH_TEST"))
	vd.gl.Call("clear", vd.gl.Get("COLOR_BUFFER_BIT"))

	return &vd, nil
}

func (vd *videoDriver) clear() {
	vd.gl.Call("clear", vd.gl.Get("COLOR_BUFFER_BIT"))
}

func (vd *videoDriver) render(frameData []uint32) error {
	return nil
}

func (vd *videoDriver) close() {

}
