package main

import (
	"syscall/js"
	"time"
)

// videoDriver provides a video driver interface with WebGL as its back end.
// This can be used inside a WebGL-capable browser as a WebAssembly
// application.
type videoDriver struct {
	targetFPS     int
	lastFrameTime time.Time
}

const framePeriod = time.Second / 60

func newVideoDriver(scaleFactor float64) (*videoDriver, error) {
	return &videoDriver{}, nil
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

	now := time.Now()
	if now.Sub(vd.lastFrameTime) < framePeriod {
		time.Sleep(framePeriod - now.Sub(vd.lastFrameTime))
	}
	vd.lastFrameTime = now

	return nil
}

func (vd *videoDriver) Close() {

}
