package main

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"
)

const (
	screenWidth  = 160
	screenHeight = 144

	screenRedrawRate = 60
)

type videoController struct {
	window   *sdl.Window
	renderer *sdl.Renderer

	env *environment
}

func newVideoController(env *environment) (videoController, error) {
	var vc videoController

	vc.env = env

	err := sdl.Init(sdl.INIT_EVERYTHING)
	if err != nil {
		return videoController{}, fmt.Errorf("initializing SDL: %v", err)
	}

	vc.window, err = sdl.CreateWindow(
		"Gopherboy",
		sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED,
		screenWidth, screenHeight,
		sdl.WINDOW_OPENGL)
	if err != nil {
		return videoController{}, fmt.Errorf("initializing window: %v", err)
	}

	vc.renderer, err = sdl.CreateRenderer(vc.window, -1, sdl.RENDERER_ACCELERATED)
	if err != nil {
		return videoController{}, fmt.Errorf("initializing renderer:", err)
	}

	return vc, nil
}

// shouldRedraw returns true if it's time for the screen to redraw, based on
// the state of the timers object.
func (vc *videoController) shouldRedraw(t timers) bool {
	// Find how many m-clock increments should happen before a redraw should
	// happen
	clocksPerRedraw := mClockRate / screenRedrawRate
	return t.mClock%clocksPerRedraw == 0
}

func (vc *videoController) update() {
	vc.renderer.SetDrawColor(255, 255, 255, 255)
	vc.renderer.Clear()

	lcdc := parseLCDC(vc.env.mbc.at(lcdcAddr))

	// Get the values of the memory registers that control the background
	// position
	scrollX := asSigned(vc.env.mbc.at(scrollXAddr))
	scrollY := asSigned(vc.env.mbc.at(scrollYAddr))

	for y := uint16(0); y < 32; y++ {
		for x := uint16(0); x < 32; x++ {
			tileAddr := lcdc.bgTileMapAddr + (y*32 + x)
			tile := vc.env.mbc.at(tileAddr)
			tileX := int(x*bgTileWidth) - int(scrollX)
			tileY := int(y*bgTileHeight) - int(scrollY)
			vc.drawTile(lcdc, tile, tileX, tileY)
		}
	}

	vc.renderer.Present()
}

func (vc *videoController) drawTile(lcdc lcdcConfig, tile uint8, tileX, tileY int) {
	// Find the address of the tile data
	var tileDataAddr uint16
	switch lcdc.windowBGTileDataTableAddr {
	case tileDataTable0:
		// Tile indexes at this data table are signed from -128 to 127
		tileDataAddr = uint16(tileDataTable0 + int(asSigned(tile))*tileBytes)
	case tileDataTable1:
		tileDataAddr = tileDataTable1 + (uint16(tile) * tileBytes)
	default:
		panic(fmt.Sprintf("unknown tile data table %#x", tileDataAddr))
	}

	tileData := vc.env.mbc.atRange(tileDataAddr, tileDataAddr+tileBytes)

	// Draw the tile on-screen
	for y := 0; y < bgTileHeight; y++ {
		lower := tileData[y*2]
		upper := tileData[(y*2)+1]

		for x := 0; x < bgTileWidth; x++ {
			// Tiles use a weird format. For each row in a tile, there are two
			// bytes. To come up with a single pixel, one bit from each byte is
			// combined into a new two-bit number which selects the color.
			lowerBit := (lower & 0x80) >> 7
			upperBit := (upper & 0x80) >> 7
			paletteID := (upperBit << 1) | lowerBit

			// Set the color based on the pixel
			if paletteID == 0x00 {
				vc.renderer.SetDrawColor(0, 0, 0, 255)
			} else if paletteID == 0x01 {
				vc.renderer.SetDrawColor(219, 179, 180, 255)
			} else if paletteID == 0x02 {
				vc.renderer.SetDrawColor(98, 78, 80, 255)
			} else if paletteID == 0x03 {
				vc.renderer.SetDrawColor(255, 255, 255, 255)
			} else {
				panic(fmt.Sprintf("Invalid palette ID %#x", paletteID))
			}

			wrappedX := tileX + x
			if wrappedX > bgWidth {
				wrappedX -= bgWidth
			} else if wrappedX < 0 {
				wrappedX += bgWidth
			}

			wrappedY := tileY + y
			if wrappedY > bgHeight {
				wrappedY -= bgHeight
			} else if wrappedY < 0 {
				wrappedY += bgHeight
			}

			vc.renderer.DrawPoint(int32(wrappedX), int32(wrappedY))

			lower <<= 1
			upper <<= 1
		}
	}
}

func (vc *videoController) destroy() {
	vc.renderer.Destroy()
	vc.window.Destroy()
}

const (
	_ uint16 = 0x0000

	// tileBytes represents the size of tile data in bytes.
	tileBytes = 16
	// bgTileWidth is the width in pixels of a background tile.
	bgTileWidth = 8
	// bgTileHeight is the height in pixels of a background tile.
	bgTileHeight = 8

	// bgWidth is the width of the background plane.
	bgWidth = 32 * bgTileWidth
	// bgHeight is the height of the background plane.
	bgHeight = 32 * bgTileHeight
)

type spriteSize string

const (
	spriteSize8x8  = "8x8"
	spriteSize8x16 = "8x16"
)

// lcdcConfig contains display configuration information as configured by the
// LCDC memory register.
type lcdcConfig struct {
	// lcdOn controls whether or not the LCD is operational.
	lcdOn bool
	// windowTileMapAddr is the address of the tile map for the window.
	windowTileMapAddr uint16
	// windowOn controls whether or not the window is being displayed.
	windowOn bool
	// windowBGTileDataTableAddr controls what tile data table should be
	// consulted for the window and the background. These two always share the
	// same data table.
	windowBGTileDataTableAddr uint16
	// bgTileMapAddr is the address of the tile map for the background.
	bgTileMapAddr uint16
	// spriteSize controls what size of sprites we're currently using.
	spriteSize spriteSize
	// spritesOn controls whether or not sprites are being displayed.
	spritesOn bool
	// windowBGOn controls whether or not the window and background is being
	// displayed.
	windowBGOn bool
}

// parseLCDC inspects the given LCDC register value for display configuration
// information.
func parseLCDC(lcdc uint8) lcdcConfig {
	var config lcdcConfig

	config.lcdOn = lcdc&0x80 == 0x80
	if lcdc&0x40 == 0x40 {
		config.windowTileMapAddr = tileMap1
	} else {
		config.windowTileMapAddr = tileMap0
	}
	config.windowOn = lcdc&0x20 == 0x20
	if lcdc&0x10 == 0x10 {
		config.windowBGTileDataTableAddr = tileDataTable1
	} else {
		config.windowBGTileDataTableAddr = tileDataTable0
	}
	if lcdc&0x08 == 0x08 {
		config.bgTileMapAddr = tileMap1
	} else {
		config.bgTileMapAddr = tileMap0
	}
	if lcdc&0x04 == 0x04 {
		config.spriteSize = spriteSize8x16
	} else {
		config.spriteSize = spriteSize8x8
	}
	config.spritesOn = lcdc&0x02 == 0x02
	config.windowBGOn = lcdc&0x01 == 0x01

	return config
}
