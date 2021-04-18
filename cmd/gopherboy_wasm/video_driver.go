package main

import (
	"fmt"
	"syscall/js"
	"time"
)

// videoDriver provides a video driver interface with WebGL as its back end.
// This can be used inside a WebGL-capable browser as a WebAssembly
// application.
type videoDriver struct {
	targetFPS     int
	lastFrameTime time.Time
	frameBuffer   js.Value
}

const framePeriod = time.Second / 50

func newVideoDriver(scaleFactor float64) (*videoDriver, error) {
	return &videoDriver{}, nil
}

func (vd *videoDriver) Render(frameData []uint8) error {
	if vd.frameBuffer.Equal(js.Undefined()) {
		fmt.Println("Initializing frame buffer...")
		vd.frameBuffer = js.Global().Get("Uint8ClampedArray").New(len(frameData))
	}
	js.CopyBytesToJS(vd.frameBuffer, frameData)

	js.Global().Call(
		"postMessage",
		[]interface{}{
			"NewFrame",
			vd.frameBuffer,
		},
	)

	now := time.Now()
	if now.Sub(vd.lastFrameTime) < framePeriod {
		time.Sleep(framePeriod - now.Sub(vd.lastFrameTime))
	}
	vd.lastFrameTime = now

	return nil
}

func (vd *videoDriver) Close() {

}
