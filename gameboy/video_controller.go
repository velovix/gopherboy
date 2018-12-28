package gameboy

import (
	"fmt"
	"time"
)

const (
	ScreenWidth  = 160
	ScreenHeight = 144

	// The number of clock ticks that the video controller has exclusive access
	// to OAM RAM for. This happens once per scan line.
	scanLineOAMClocks = 80
	// The number of clock ticks that the video controller has exclusive access
	// to VRAM for. This happens once per scan line.
	scanLineVRAMClocks = 172
	// The number of clock ticks in between scan lines.
	horizontalBlankClocks = 204
	// The amount of clocks taken for a scan line.
	scanLineFullClocks = scanLineOAMClocks + scanLineVRAMClocks + horizontalBlankClocks
	// The number of clock ticks in between frame draws.
	verticalBlankClocks = 4560
	// The total number of clocks taken for a frame.
	fullFrameClocks = (scanLineFullClocks * ScreenHeight) + verticalBlankClocks

	// spriteWidth is the pixel width of a sprite
	spriteWidth = 8
	// spriteTallHeight is the pixel height of a sprite if 8x16 mode is
	// enabled.
	spriteTallHeight = 16
	// spriteShortHeight is the pixel height of a sprite if 8x8 mode is
	// enabled.
	spriteShortHeight = 8

	// targetFPS is the FPS of the Game Boy screen.
	targetFPS = 60
)

type drawStep int

const (
	_ drawStep = 0
	oamStep
	vramStep
	horizontalBlankStep
	verticalBlankStep
)

type videoController struct {
	driver VideoDriver

	frameTick      int
	drawnScanLines int
	lcdc           lcdcConfig
	// All OAM entries. Used to control sprites on-screen.
	oams []oam
	// scrollX controls the X position of the background.
	scrollX int8
	// scrollY controls the Y position of the background.
	scrollY int8
	// windowX controls the X position of the window in screen coordinates.
	windowX uint8
	// windowY controls the Y position of the window is screen coordinates.
	windowY uint8
	// bgPalette is the palette for the background.
	bgPalette map[uint8]color
	// spritePalette0 is the first of the two available sprite palettes.
	spritePalette0 map[uint8]color
	// spritePalette1 is the second of the two available sprite palettes.
	spritePalette1 map[uint8]color

	// Raw frame data in 8-bit RGBA format.
	currFrame []uint8

	timers *timers
	state  *State

	// Used for finding FPS
	lastSecond time.Time
	frameCnt   int

	// Used to cap FPS
	lastFrameTime time.Time
	// If true, frame rate will not be capped.
	unlimitedFPS bool
}

func newVideoController(state *State, timers *timers, driver VideoDriver) *videoController {
	var vc videoController

	vc.driver = driver

	vc.state = state
	vc.timers = timers

	vc.lastSecond = time.Now()

	return &vc
}

// tick progresses the video controller by the given number of cycles.
func (vc *videoController) tick(opTime int) {
	// Check if the LCD should be on
	if !vc.loadLCDC().lcdOn {
		return
	}

	for i := 0; i < opTime; i++ {
		if vc.frameTick == 0 {
			// Get ready for a new frame draw
			vc.driver.Clear()
			// Read some frame-wide values
			vc.scrollY = asSigned(vc.state.mmu.at(scrollYAddr))
			vc.windowY = vc.state.mmu.at(windowPosYAddr)

			// Make a new frame
			vc.currFrame = make([]uint8, ScreenWidth*ScreenHeight*4)
		}

		// Update the LY register with the current scan line. Note that this
		// value increments even during VBlank even though new scan lines
		// aren't actually being drawn.
		currScanLine := vc.frameTick / scanLineFullClocks
		vc.state.mmu.setNoNotify(lyAddr, uint8(currScanLine))

		if vc.frameTick < scanLineFullClocks*ScreenHeight {
			// We're still drawing scan lines
			scanLineProgress := vc.frameTick % scanLineFullClocks

			switch scanLineProgress {
			case 0:
				// We're in mode 2, OAM read mode.
				vc.setMode(vcMode2)
				// TODO(velovix): Lock OAM?

				vc.oams = vc.loadOAM()

				// This is the start of this scan line, read some scan line
				// wide values
				vc.lcdc = vc.loadLCDC()
				vc.scrollX = asSigned(vc.state.mmu.at(scrollXAddr))
				vc.windowX = vc.state.mmu.at(windowPosXAddr)
			case scanLineOAMClocks:
				// We're in mode 3, OAM and VRAM transfer mode.
				vc.setMode(vcMode3)

				vc.bgPalette = vc.loadBGPalette()
				vc.spritePalette0 = vc.loadSpritePalette(0)
				vc.spritePalette1 = vc.loadSpritePalette(1)
				// TODO(velovix): Lock VRAM
			case scanLineVRAMClocks:
				// We're in mode 0, HBlank period
				vc.setMode(vcMode0)
				// TODO(velovix): Unlock things
				// We're ready to draw the scan line
				vc.drawScanLine(uint8(currScanLine))
			}
		} else {
			// We're in mode 1, VBlank period
			vc.setMode(vcMode1)

			if vc.frameTick == scanLineFullClocks*ScreenHeight {
				// We just finished drawing the frame
				vblankInterruptEnabled := vc.state.mmu.at(ieAddr)&0x01 == 0x01
				if vc.state.interruptsEnabled && vblankInterruptEnabled {
					vc.state.mmu.setNoNotify(ifAddr, vc.state.mmu.at(ifAddr)|0x01)
				}

				vc.driver.Render(vc.currFrame)

				if !vc.unlimitedFPS {
					if time.Since(vc.lastFrameTime) < time.Second/targetFPS {
						time.Sleep((time.Second / targetFPS) - time.Since(vc.lastFrameTime))
					}
				}
				vc.lastFrameTime = time.Now()

				vc.frameCnt++
				if time.Since(vc.lastSecond) >= time.Second {
					fmt.Println("FPS:", vc.frameCnt)
					vc.frameCnt = 0
					vc.lastSecond = time.Now()
				}
			}
		}

		// Should this be before or after operations
		vc.frameTick++
		if vc.frameTick == fullFrameClocks {
			vc.frameTick = 0
		}

	}
}

// drawScanLine draws a scan line at the given height position.
func (vc *videoController) drawScanLine(line uint8) {
	for x := uint8(0); x < ScreenWidth; x++ {
		bgDotCode := vc.bgDotCode(x, line)

		var pixelColor color
		oamEntry, hasSprite := vc.spriteAt(x, line)
		if hasSprite && vc.lcdc.spritesOn {
			// If the sprite has priority 1 and the background dot data is
			// other than zero, the background will be shown "over" the sprite
			if oamEntry.priority && bgDotCode != 0 {
				pixelColor = vc.bgPalette[bgDotCode]
			} else {
				// Get the color at this specific place on the sprite
				xOffset := x + spriteWidth - oamEntry.xPos
				yOffset := line + spriteTallHeight - oamEntry.yPos

				if oamEntry.xFlip {
					// Flip the sprite horizontally
					xOffset = (spriteWidth - 1) - xOffset
				}
				if oamEntry.yFlip {
					// Flip the sprite vertically
					// TODO(velovix): This will have to be readdressed when
					// tall sprite support is added
					switch vc.lcdc.spriteSize {
					case spriteSize8x8:
						yOffset = (spriteShortHeight - 1) - yOffset
					case spriteSize8x16:
						yOffset = (spriteTallHeight - 1) - yOffset
					}
				}

				spriteDotCode := vc.dotCodeInSprite(
					oamEntry.spriteNumber, int(xOffset), int(yOffset))

				if spriteDotCode == 0 {
					// As a special case, if the dot code is zero the sprite is
					// always transparent, regardless of the palette
					pixelColor = vc.bgPalette[bgDotCode]
				} else {
					// Use the selected sprite palette
					if oamEntry.paletteNumber == 0 {
						pixelColor = vc.spritePalette0[spriteDotCode]
					} else if oamEntry.paletteNumber == 1 {
						pixelColor = vc.spritePalette1[spriteDotCode]
					} else {
						panic(fmt.Sprintf("unknown sprite palette value %v",
							oamEntry.paletteNumber))
					}
				}
			}
		} else if vc.lcdc.windowBGOn && vc.lcdc.windowOn && vc.coordInWindow(x, line) {
			dotCode := vc.windowDotCode(x, line)
			pixelColor = vc.bgPalette[dotCode]
		} else if vc.lcdc.windowBGOn {
			pixelColor = vc.bgPalette[bgDotCode]
		}

		// Add this pixel to the in-progress frame
		pixelStart := ((int(line) * ScreenWidth) + int(x)) * 4
		vc.currFrame[pixelStart] = pixelColor.r
		vc.currFrame[pixelStart+1] = pixelColor.g
		vc.currFrame[pixelStart+2] = pixelColor.b
		vc.currFrame[pixelStart+3] = pixelColor.a
	}
}

// coordInWindow returns true if the given coordinates are in the window's
// current area.
//
// The coordinate system for the window is a little bit funky. For whatever
// reason, the top left of the screen is actually at windowX=7, not windowX=0.
func (vc *videoController) coordInWindow(x, y uint8) bool {
	return x >= vc.windowX-7 && y >= vc.windowY
}

// spriteAt returns the sprite that is at the given X and Y value. If none
// exists, an empty OAM and an ok value of false will be returned.
func (vc *videoController) spriteAt(x, y uint8) (entry oam, ok bool) {
	for _, entry := range vc.oams {
		var spriteTop uint8
		switch vc.lcdc.spriteSize {
		case spriteSize8x8:
			spriteTop = entry.yPos - spriteShortHeight
		case spriteSize8x16:
			spriteTop = entry.yPos
		}

		// Check if the sprite this OAM entry corresponds to is in the given
		// point. Remember that a sprite's X and Y position is relative to the
		// bottom right of the sprite.
		if x < entry.xPos && int(x) >= int(entry.xPos)-spriteWidth &&
			y < spriteTop && int(y) >= int(entry.yPos)-spriteTallHeight {
			return entry, true
		}
	}

	return oam{}, false
}

// bgDotCode returns the dot code in the background layer at the given screen
// coordinates.
func (vc *videoController) bgDotCode(x, y uint8) uint8 {
	// Get the coordinates relative to the background and wrap them if
	// necessary
	bgX := int(x) + int(vc.scrollX)
	if bgX < 0 {
		bgX += bgWidth
	} else if bgX > bgWidth {
		bgX -= bgWidth
	}
	bgY := int(y) + int(vc.scrollY)
	if bgY < 0 {
		bgY += bgHeight
	} else if bgY > bgHeight {
		bgY -= bgHeight
	}

	// Get the tile this point is inside of
	tileOffset := (bgY/bgTileHeight)*bgWidthInTiles + (bgX / bgTileWidth)
	tileAddr := vc.lcdc.bgTileMapAddr + uint16(tileOffset)
	tile := vc.state.mmu.at(tileAddr)

	// Find the dot code at this specific place in the tile
	inTileX := bgX % bgTileWidth
	inTileY := bgY % bgTileHeight

	return vc.dotCodeInTile(tile, inTileX, inTileY)
}

// windowDotCode returns the dot code in the window layer at the given screen
// coordinates. The coordinates should be checked to see if they are in the
// window before calling this method.
func (vc *videoController) windowDotCode(x, y uint8) uint8 {
	if !vc.coordInWindow(x, y) {
		panic("Attempt to load a window dot code in a non-window location")
	}

	// Get the x and y coordinates in window space
	winX := int(x - vc.windowX + 7)
	winY := int(y - vc.windowY)

	tileOffset := (winY/windowTileHeight)*windowWidthInTiles + (winX / windowTileWidth)
	tileAddr := vc.lcdc.windowTileMapAddr + uint16(tileOffset)
	tile := vc.state.mmu.at(tileAddr)

	inTileX := winX % windowTileWidth
	inTileY := winY % windowTileHeight

	return vc.dotCodeInTile(tile, inTileX, inTileY)
}

// dotCodeInTile finds the dot code for a place in a tile given the tile's ID
// and the coordinates within the tile to look at.
func (vc *videoController) dotCodeInTile(tileID uint8, inTileX, inTileY int) uint8 {
	// Find the address of the tile data
	var tileDataAddr uint16
	switch vc.lcdc.windowBGTileDataTableAddr {
	case tileDataTable0:
		// Tile indexes at this data table are signed from -128 to 127
		tileDataAddr = uint16(tileDataTable0 + int(int8(tileID))*tileBytes)
	case tileDataTable1:
		tileDataAddr = tileDataTable1 + (uint16(tileID) * tileBytes)
	default:
		panic(fmt.Sprintf("unknown tile data table %#x", tileDataAddr))
	}

	lower := vc.state.mmu.at(tileDataAddr + uint16(inTileY*2))
	upper := vc.state.mmu.at(tileDataAddr + uint16((inTileY*2)+1))

	lower <<= uint(inTileX)
	upper <<= uint(inTileX)

	lowerBit := (lower & 0x80) >> 7
	upperBit := (upper & 0x80) >> 7
	return (upperBit << 1) | lowerBit
}

// dotCodeInSprite finds the dot code for a place in a sprite given the
// sprite's ID and the coordinates within the sprite to look at.
func (vc *videoController) dotCodeInSprite(spriteID uint8, inSpriteX, inSpriteY int) uint8 {
	if vc.lcdc.spriteSize == spriteSize8x16 {
		// The first bit of the sprite ID is ignored in this mode. This is
		// because sprites in this mode take up twice the space, making only
		// every other sprite ID valid
		spriteID &= ^uint8(0x1)
	}

	// Find the address of the tile data
	spriteDataAddr := spriteDataTable + uint16(spriteID)*spriteBytes8x8

	lower := vc.state.mmu.at(spriteDataAddr + uint16(inSpriteY*2))
	upper := vc.state.mmu.at(spriteDataAddr + uint16((inSpriteY*2)+1))

	lower <<= uint(inSpriteX)
	upper <<= uint(inSpriteX)

	lowerBit := (lower & 0x80) >> 7
	upperBit := (upper & 0x80) >> 7
	return (upperBit << 1) | lowerBit
}

func (vc *videoController) destroy() {
	vc.driver.Close()
}

// setMode updates the necessary registers to show what mode the video
// controller is in.
func (vc *videoController) setMode(mode vcMode) {
	statVal := vc.state.mmu.at(statAddr)

	// Clear the current mode value
	statVal &= 0xFC
	// Set the mode
	statVal |= uint8(mode)

	vc.state.mmu.setNoNotify(statAddr, statVal)
}

const (
	_ uint16 = 0x0000

	// tileBytes represents the size of tile data in bytes.
	tileBytes = 16
	// bgTileWidth is the width in pixels of a background tile.
	bgTileWidth = 8
	// bgTileHeight is the height in pixels of a background tile.
	bgTileHeight = 8
	// windowTileWidth is the width in pixels of a window tile.
	windowTileWidth = bgTileWidth
	// windowTileHeight is the height in pixels of a window tile.
	windowTileHeight = bgTileHeight

	spriteBytes8x8  = 16
	spriteBytes8x16 = 32

	// bgWidthInTiles is the number of tiles per row in the background.
	bgWidthInTiles = 32
	// bgHeightInTiles is the number of tiles per column in the background.
	bgHeightInTiles = 32

	// windowWidthInTiles is the number of tiles per row in the window.
	windowWidthInTiles = bgWidthInTiles
	// windowHeightInTiles is the number of tiles per column in the window.
	windowHeightInTiles = bgHeightInTiles

	// bgWidth is the width of the background plane.
	bgWidth = bgWidthInTiles * bgTileWidth
	// bgHeight is the height of the background plane.
	bgHeight = 32 * bgTileHeight
)

type vcMode uint8

const (
	// The HBlank period mode.
	vcMode0 vcMode = 0
	// The VBlank period mode.
	vcMode1
	// The OAM RAM loading mode. OAM RAM may not be written to at this time.
	vcMode2
	// The VRAM and OAM RAM loading mode. VRAM and OAM RAM may not be written
	// to at this time.
	vcMode3
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

// loadLCDC inspects the LCDC register value for display configuration
// information.
func (vc *videoController) loadLCDC() lcdcConfig {
	var config lcdcConfig
	lcdc := vc.state.mmu.at(lcdcAddr)

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

// oam represents a single OAM entry. OAM stands for Object Attribute Memory
// and are used to control sprites on screen.
//
// A single OAM entry is 4 bytes in size.
//
// Byte 0: Y position on-screen
// Byte 1: X position on-screen, on the right of the sprite
// Byte 2: The sprite/tile number from 0-255. This controls what the sprite
//         looks like.
// Byte 3: Flags controlling other attributes of the sprite.
//     Bit 7: Priority. If 1, the sprite will be displayed under all background
//            pixels except ones with dot data equal to zero. If 0, the sprite
//            will be drawn over all background pixels.
//     Bit 6: Y Flip. If 1, the sprite will be flipped vertically.
//     Bit 5: X Flip. If 1, the sprite will be flipped horizontally.
//     Bit 4: Palette number. If 1, the sprite will use object palette 1, if 0,
//            the sprite will use object palette 0.
//     Bits 3-0: Unused
type oam struct {
	yPos          uint8
	xPos          uint8
	spriteNumber  uint8
	priority      bool
	yFlip         bool
	xFlip         bool
	paletteNumber uint8
}

// loadOAM loads all OAM entries from memory.
func (vc *videoController) loadOAM() []oam {
	oams := make([]oam, 40)

	for i := 0; i < 40; i++ {
		entryStart := uint16(oamRAMAddr + (i * oamBytes))
		oams[i] = oam{
			yPos:         vc.state.mmu.at(entryStart),
			xPos:         vc.state.mmu.at(entryStart + 1),
			spriteNumber: vc.state.mmu.at(entryStart + 2),
		}

		flags := vc.state.mmu.at(entryStart + 3)
		oams[i].priority = flags&0x80 == 0x80
		oams[i].yFlip = flags&0x40 == 0x40
		oams[i].xFlip = flags&0x20 == 0x20

		if flags&0x10 == 0x10 {
			oams[i].paletteNumber = 1
		} else {
			oams[i].paletteNumber = 0
		}
	}

	return oams
}

const (
	// The size in bytes of an OAM entry.
	oamBytes = 4
)

// statConfig configures LCD configuration information as configured by the
// STAT memory register.
type statConfig struct {
	// lyEqualsLYCInterruptOn is true if an interrupt should be generated when
	// the LY and LYC memory registers are equal.
	lyEqualsLYCInterruptOn bool
	// mode2InterruptOn is true if an interrupt should be generated when the
	// video controller switches to mode 2.
	mode2InterruptOn bool
	// mode1InterruptOn is true if an interrupt should be generated when the
	// video controller switches to mode 1.
	mode1InterruptOn bool
	// mode0InterruptOn is true if an interrupt should be generated when the
	// video controller switches to mode 0.
	mode0InterruptOn bool
	// lyEqualsLYC is true if the LY and LYC memory registers are equal.
	lyEqualsLYC bool
	// vcMode is the current mode of the video controller.
	mode vcMode
}

// loadSTAT inspects the STAT register value for LCD configuration information.
func (vc *videoController) loadSTAT() statConfig {
	var config statConfig
	stat := vc.state.mmu.at(statAddr)

	config.lyEqualsLYCInterruptOn = stat&0x40 == 0x40
	config.mode2InterruptOn = stat&0x20 == 0x20
	config.mode1InterruptOn = stat&0x10 == 0x10
	config.mode0InterruptOn = stat&0x08 == 0x08
	config.lyEqualsLYC = stat&0x04 == 0x04
	config.mode = vcMode(stat & 0x03)

	return config
}

// saveSTAT saves the given STAT configuration into the memory register.
func (vc *videoController) saveSTAT(config statConfig) {
	statVal := vc.state.mmu.at(statAddr)

	if config.lyEqualsLYCInterruptOn {
		statVal |= 0x40
	} else {
		statVal &= ^uint8(0x40)
	}
	if config.mode2InterruptOn {
		statVal |= 0x20
	} else {
		statVal &= ^uint8(0x20)
	}
	if config.mode1InterruptOn {
		statVal |= 0x10
	} else {
		statVal &= ^uint8(0x10)
	}
	if config.mode0InterruptOn {
		statVal |= 0x08
	} else {
		statVal &= ^uint8(0x08)
	}
	if config.lyEqualsLYC {
		statVal |= 0x04
	} else {
		statVal &= ^uint8(0x04)
	}
	// Clear and set the mode
	statVal &= 0xFC
	statVal |= uint8(config.mode)

	vc.state.mmu.set(statAddr, statVal)
}

// loadBGPalette inspects the BGP register value and returns a map that maps
// dot data to actual colors.
func (vc *videoController) loadBGPalette() map[uint8]color {
	bgp := vc.state.mmu.at(bgpAddr)
	palette := make(map[uint8]color)

	for dotData := uint8(0); dotData <= 0x03; dotData++ {
		paletteOption := bgp & 0x03

		var c color
		switch paletteOption {
		case 0x00:
			//c = color{52, 104, 86, 255}
			c = color{224, 248, 208, 255}
		case 0x01:
			//c = color{8, 24, 32, 255}
			c = color{136, 192, 112, 255}
		case 0x02:
			//c = color{52, 104, 86, 255}
			c = color{52, 104, 86, 255}
		case 0x03:
			//c = color{8, 24, 32, 255}
			c = color{8, 24, 32, 255}
		}

		palette[dotData] = c
		bgp >>= 2
	}

	return palette
}

// loadSpritePalette inspects the OBP0 or OPB1 register values and returns a
// map that maps dot data to actual colors.
func (vc *videoController) loadSpritePalette(paletteNum int) map[uint8]color {
	obp := vc.state.mmu.at(opb0Addr + uint16(paletteNum))
	palette := make(map[uint8]color)

	for dotData := uint8(0); dotData <= 0x03; dotData++ {
		paletteOption := obp & 0x03

		var c color
		switch paletteOption {
		case 0x00:
			//c = color{52, 104, 86, 255}
			c = color{224, 248, 208, 255}
		case 0x01:
			//c = color{8, 24, 32, 255}
			c = color{136, 192, 112, 255}
		case 0x02:
			//c = color{52, 104, 86, 255}
			c = color{52, 104, 86, 255}
		case 0x03:
			//c = color{8, 24, 32, 255}
			c = color{8, 24, 32, 255}
		}

		palette[dotData] = c
		obp >>= 2
	}

	return palette
}

type color struct {
	r, g, b, a uint8
}
