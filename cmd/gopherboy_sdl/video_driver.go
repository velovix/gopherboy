package main

import (
	"fmt"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/velovix/gopherboy/gameboy"
)

// videoDriver provides a video driver interface with SDL as its back end. This
// can be used on most platforms as a native application.
type videoDriver struct {
	window   *sdl.Window
	renderer *sdl.Renderer
}

// newVideoDriver creates a new SDL video driver. The scale factor resizes the
// window by that value.
func newVideoDriver(scaleFactor float64) (*videoDriver, error) {
	var vd videoDriver

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, fmt.Errorf("initializing SDL: %v", err)
	}

	vd.window, err = sdl.CreateWindow(
		"Gopherboy",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(gameboy.ScreenWidth*scaleFactor),
		int32(gameboy.ScreenHeight*scaleFactor),
		sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, fmt.Errorf("initializing window: %v", err)
	}

	vd.renderer, err = sdl.CreateRenderer(vd.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, fmt.Errorf("initializing renderer: %v", err)
	}

	vd.renderer.SetDrawColor(255, 255, 255, 255)

	return &vd, nil
}

// clear clears the screen in preparation for a new frame.
func (vd *videoDriver) Clear() {
	vd.renderer.Clear()
}

// render renders the given RGBA frame data on-screen. This is done by turning
// it into a texture and copying it onto the renderer.
func (vd *videoDriver) Render(frameData []uint32) error {
	surface, err := sdl.CreateRGBSurfaceFrom(
		unsafe.Pointer(&frameData[0]),
		gameboy.ScreenWidth,
		gameboy.ScreenHeight,
		32,                    // Bits per pixel
		4*gameboy.ScreenWidth, // Bytes per row
		0xFF000000,            // Bitmask for R value
		0x00FF0000,            // Bitmask for G value
		0x0000FF00,            // Bitmask for B value
		0x000000FF,            // Bitmask for alpha value
	)
	if err != nil {
		return fmt.Errorf("creating surface: %v", err)
	}
	defer surface.Free()

	texture, err := vd.renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return fmt.Errorf("converting surface to a texture: %v", err)
	}
	defer texture.Destroy()

	err = vd.renderer.Copy(texture, nil, nil)
	if err != nil {
		return fmt.Errorf("copying frame to screen: %v", err)
	}

	vd.renderer.Present()

	return nil
}

// close de-initializes the video driver in preparation for exit.
func (vd *videoDriver) Close() {
	vd.renderer.Destroy()
	vd.window.Destroy()
}
