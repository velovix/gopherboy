package main

import (
	"time"
	"unsafe"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/velovix/gopherboy/gameboy"
	"golang.org/x/xerrors"
)

// targetFPS is the FPS of the Game Boy screen.
const targetFPS = 60

// videoDriver provides a video driver interface with SDL as its back end. This
// can be used on most platforms as a native application.
type videoDriver struct {
	window   *sdl.Window
	renderer *sdl.Renderer

	unlimitedFPS  bool
	lastFrameTime time.Time

	readyForNewFrame bool
}

// newVideoDriver creates a new SDL video driver. The scale factor resizes the
// window by that value.
func newVideoDriver(scaleFactor float64, unlimitedFPS bool) (*videoDriver, error) {
	var vd videoDriver

	vd.unlimitedFPS = unlimitedFPS

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return nil, xerrors.Errorf("initializing SDL: %v", err)
	}

	vd.window, err = sdl.CreateWindow(
		"Gopherboy",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		int32(gameboy.ScreenWidth*scaleFactor),
		int32(gameboy.ScreenHeight*scaleFactor),
		sdl.WINDOW_OPENGL)
	if err != nil {
		return nil, xerrors.Errorf("initializing window: %v", err)
	}

	vd.renderer, err = sdl.CreateRenderer(vd.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return nil, xerrors.Errorf("initializing renderer: %v", err)
	}

	vd.renderer.SetDrawColor(255, 255, 255, 255)

	vd.readyForNewFrame = true

	return &vd, nil
}

// Render renders the given RGBA frame data on-screen. This is done by turning
// it into a texture and copying it onto the renderer.
func (vd *videoDriver) Render(frameData []uint8) error {
	if !vd.readyForNewFrame {
		return nil
	}

	vd.readyForNewFrame = false

	doOnMainThread(func() {
		surface, err := sdl.CreateRGBSurfaceFrom(
			unsafe.Pointer(&frameData[0]),
			gameboy.ScreenWidth,
			gameboy.ScreenHeight,
			32,                    // Bits per pixel
			4*gameboy.ScreenWidth, // Bytes per row
			0x000000FF,            // Bitmask for R value
			0x0000FF00,            // Bitmask for G value
			0x00FF0000,            // Bitmask for B value
			0xFF000000,            // Bitmask for alpha value
		)
		if err != nil {
			err = xerrors.Errorf("creating surface: %w", err)
			return
		}
		defer surface.Free()

		texture, err := vd.renderer.CreateTextureFromSurface(surface)
		if err != nil {
			err = xerrors.Errorf("converting surface to a texture: %w", err)
			return
		}
		defer texture.Destroy()

		err = vd.renderer.Copy(texture, nil, nil)
		if err != nil {
			err = xerrors.Errorf("copying frame to screen: %w", err)
			return
		}

		vd.renderer.Present()

		if !vd.unlimitedFPS {
			if time.Since(vd.lastFrameTime) < time.Second/targetFPS {
				time.Sleep((time.Second / targetFPS) - time.Since(vd.lastFrameTime))
			}
			vd.lastFrameTime = time.Now()
		}

		vd.readyForNewFrame = true
	}, true)

	return nil
}

// close de-initializes the video driver in preparation for exit.
func (vd *videoDriver) Close() {
	doOnMainThread(func() {
		vd.renderer.Destroy()
		vd.window.Destroy()
	}, false)
}
