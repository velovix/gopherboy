package gameboy

import (
	"fmt"
	"sort"
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

	// maxSpritesPerScanLine is the maximum amount of sprites that can share a
	// row. Any more sprites will not be drawn.
	maxSpritesPerScanLine = 10

	// maxOAMEntries is the maximum amount of OAM entries that can fit in OAM
	// memory.
	maxOAMEntries = 40
)

type drawStep int

const (
	_ drawStep = 0
	oamStep
	vramStep
	horizontalBlankStep
	verticalBlankStep
)

// videoController emulates the Game Boy's video hardware. It produces frames
// that may be displayed on screen.
type videoController struct {
	driver VideoDriver

	// If false, the screen is off and no draw operations happen.
	lcdOn bool

	frameTick      int
	drawnScanLines int
	lcdc           lcdcConfig
	// spritesOnScanLine is a list of up to 10 sprites that are on the scan
	// line that is currently being drawn.
	spritesOnScanLine []oam
	// spriteCount is the number of sprites that were found on the current scan
	// line. This is the number of usable OAM entries in spritesOnScanLine.
	spriteCount int
	// scrollX controls the X position of the background.
	scrollX int8
	// scrollY controls the Y position of the background.
	scrollY int8
	// windowX controls the X position of the window in screen coordinates.
	windowX uint8
	// windowY controls the Y position of the window is screen coordinates.
	windowY uint8
	// bgPalette is the palette for the background. It maps dot data to its
	// corresponding color.
	bgPalette [4]color
	// spritePalette0 is the first of the two available sprite palettes. It
	// maps dot data to its corresponding color.
	spritePalette0 [4]color
	// spritePalette1 is the second of the two available sprite palettes. It
	// maps dot data to its corresponding color.
	spritePalette1 [4]color

	// Raw frame data in 8-bit RGBA format.
	currFrame []uint8

	state *State

	// Used for finding FPS
	lastSecond    time.Time
	frameCnt      int
	fpsQueryCount int
	fpsTotal      int
}

func newVideoController(state *State, driver VideoDriver) *videoController {
	vc := &videoController{
		lcdOn:             true,
		driver:            driver,
		state:             state,
		lastSecond:        time.Now(),
		spritesOnScanLine: make([]oam, maxSpritesPerScanLine),
		currFrame:         make([]uint8, ScreenWidth*ScreenHeight*4),
	}

	vc.currFrame = make([]uint8, ScreenWidth*ScreenHeight*4)

	vc.state.mmu.subscribeTo(statAddr, vc.onSTATWrite)
	vc.state.mmu.subscribeTo(lcdcAddr, vc.onLCDCWrite)

	return vc
}

// tick progresses the video controller by the given number of cycles.
func (vc *videoController) tick(opTime int) {
	if !vc.lcdOn {
		return
	}

	stat := vc.loadSTAT()

	// Check which interrupts are currently enabled
	lcdcInterruptsEnabled := vc.state.mmu.memory[ieAddr]&0x02 == 0x02
	lyEqualsLYCInterruptEnabled := stat.lyEqualsLYCInterruptOn &&
		lcdcInterruptsEnabled &&
		vc.state.interruptsEnabled
	mode2InterruptEnabled := stat.mode2InterruptOn &&
		lcdcInterruptsEnabled &&
		vc.state.interruptsEnabled
	mode1InterruptEnabled := stat.mode1InterruptOn &&
		lcdcInterruptsEnabled &&
		vc.state.interruptsEnabled
	mode0InterruptEnabled := stat.mode0InterruptOn &&
		lcdcInterruptsEnabled &&
		vc.state.interruptsEnabled

	for i := 0; i < opTime; i++ {
		if vc.frameTick == 0 {
			// Read some frame-wide values
			vc.scrollY = int8(vc.state.mmu.memory[scrollYAddr])
			vc.windowY = vc.state.mmu.memory[windowPosYAddr]
		}

		// Update the LY register with the current scan line. Note that this
		// value increments even during VBlank even though new scan lines
		// aren't actually being drawn.
		currScanLine := (vc.frameTick / scanLineFullClocks)
		// TODO(velovix): Is adding a 1 to this correct behavior? Not adding 1
		// results in visual glitche
		vc.state.mmu.memory[lyAddr] = uint8(currScanLine + 1)

		// Check if LY==LYC
		lyJustChanged := vc.frameTick%scanLineFullClocks == 0
		if lyJustChanged {
			lyc := vc.state.mmu.memory[lyAddr]
			// TODO(velovix): Why do I need to add a one here?
			stat.lyEqualsLYC = (currScanLine + 1) == int(lyc)
			// Trigger an interrupt if they're equal and the interrupt is
			// enabled
			if stat.lyEqualsLYC && lyEqualsLYCInterruptEnabled {
				vc.state.mmu.memory[ifAddr] = vc.state.mmu.memory[ifAddr] | 0x02
			}
		}

		if vc.frameTick < scanLineFullClocks*ScreenHeight {
			// We're still drawing scan lines
			scanLineProgress := vc.frameTick % scanLineFullClocks

			switch scanLineProgress {
			case 0:
				// We're in mode 2, OAM read mode.
				stat.mode = vcMode2
				if mode2InterruptEnabled {
					vc.state.mmu.memory[ifAddr] = vc.state.mmu.memory[ifAddr] | 0x02
				}
				// TODO(velovix): Lock OAM?

				vc.loadSpritesOnScanLine(uint8(currScanLine))

				// This is the start of this scan line, read some scan line
				// wide values
				vc.lcdc = vc.loadLCDC()
				vc.scrollX = int8(vc.state.mmu.memory[scrollXAddr])
				vc.windowX = vc.state.mmu.memory[windowPosXAddr]
			case scanLineOAMClocks:
				// We're in mode 3, OAM and VRAM transfer mode.
				stat.mode = vcMode3

				vc.loadBGPalette()
				vc.loadSpritePalettes()
				// TODO(velovix): Lock VRAM
			case scanLineVRAMClocks:
				// We're in mode 0, HBlank period
				stat.mode = vcMode0
				if mode0InterruptEnabled {
					vc.state.mmu.memory[ifAddr] = vc.state.mmu.memory[ifAddr] | 0x02
				}

				// TODO(velovix): Unlock things
				vc.drawScanLine(uint8(currScanLine))
			}
		} else {
			// We're in mode 1, VBlank period
			stat.mode = vcMode1
			if mode1InterruptEnabled {
				vc.state.mmu.memory[ifAddr] = vc.state.mmu.memory[ifAddr] | 0x02
			}

			if vc.frameTick == scanLineFullClocks*ScreenHeight {
				// We just finished drawing the frame
				vblankInterruptEnabled := vc.state.mmu.memory[ieAddr]&0x01 == 0x01
				if vc.state.interruptsEnabled && vblankInterruptEnabled {
					vc.state.mmu.memory[ifAddr] = vc.state.mmu.memory[ifAddr] | 0x01
				}

				vc.driver.Render(vc.currFrame)

				vc.frameCnt++
				if time.Since(vc.lastSecond) >= time.Second {
					vc.fpsQueryCount++
					vc.fpsTotal += vc.frameCnt

					fmt.Println("Average FPS:", vc.fpsTotal/vc.fpsQueryCount)
					fmt.Println("FPS:", vc.frameCnt)

					vc.frameCnt = 0
					vc.lastSecond = time.Now()

				}
			}
		}

		vc.frameTick++
		if vc.frameTick == fullFrameClocks {
			vc.frameTick = 0
		}
	}

	vc.saveSTAT(stat)
}

// drawScanLine draws a scan line at the given height position.
func (vc *videoController) drawScanLine(line uint8) {
	for x := uint8(0); x < ScreenWidth; x++ {
		bgDotCode := vc.bgDotCode(x, line)

		var pixelColor color
		pixelDrawn := false

		if vc.lcdc.spritesOn {
			// Look for a sprite to draw at this position
			for _, sprite := range vc.spritesAt(x) {
				// If the sprite has priority 1 and the background dot data is
				// other than zero, that part of the sprite will not be drawn
				// and the background will be seen instead, assuming it is
				// enabled.
				if sprite.priority && bgDotCode != 0 {
					continue
				}

				// Get the color at this specific place on the sprite
				xOffset := x + spriteWidth - sprite.xPos
				yOffset := line + spriteTallHeight - sprite.yPos

				if sprite.xFlip {
					// Flip the sprite horizontally
					xOffset = (spriteWidth - 1) - xOffset
				}
				if sprite.yFlip {
					// Flip the sprite vertically
					switch vc.lcdc.spriteSize {
					case spriteSize8x8:
						yOffset = (spriteShortHeight - 1) - yOffset
					case spriteSize8x16:
						yOffset = (spriteTallHeight - 1) - yOffset
					}
				}

				spriteDotCode := vc.dotCodeInSprite(
					sprite.spriteNumber, int(xOffset), int(yOffset))

				// The dot code zero in sprites represents transparency
				if spriteDotCode != 0 {
					// Use the selected sprite palette
					pixelDrawn = true
					if sprite.paletteNumber == 0 {
						pixelColor = vc.spritePalette0[spriteDotCode]
					} else if sprite.paletteNumber == 1 {
						pixelColor = vc.spritePalette1[spriteDotCode]
					} else {
						panic(fmt.Sprintf("unknown sprite palette value %v",
							sprite.paletteNumber))
					}

					// We've already found our sprite to draw. Lower priority
					// sprites at this position will not be drawn.
					break
				}
			}
		}

		// Draw the window if a sprite hasn't already been drawn
		if !pixelDrawn && vc.lcdc.windowBGOn && vc.lcdc.windowOn && vc.coordInWindow(x, line) {
			dotCode := vc.windowDotCode(x, line)
			pixelColor = vc.bgPalette[dotCode]
			pixelDrawn = true
		}

		// Draw the background if a sprite or window hasn't already been drawn
		if !pixelDrawn && vc.lcdc.windowBGOn {
			pixelColor = vc.bgPalette[bgDotCode]
			pixelDrawn = true
		}

		// Add this pixel to the in-progress frame
		pixelStart := ((int(line) * ScreenWidth) + int(x)) * 4
		vc.currFrame[pixelStart] = pixelColor.r
		vc.currFrame[pixelStart+1] = pixelColor.g
		vc.currFrame[pixelStart+2] = pixelColor.b
		vc.currFrame[pixelStart+3] = pixelColor.a
	}
}

// onStatWrite is called when the STAT register is written to.
func (vc *videoController) onSTATWrite(addr uint16, val uint8) uint8 {
	// TODO(velovix): Consider only reloading this register when this method is
	// called as a performance optimization
	// The 7th bit of the register is unused
	return val | 0x80
}

// coordInWindow returns true if the given coordinates are in the window's
// current area.
//
// The coordinate system for the window is a little bit funky. For whatever
// reason, the top left of the screen is actually at windowX=7, not windowX=0.
func (vc *videoController) coordInWindow(x, y uint8) bool {
	return x >= vc.windowX-7 && y >= vc.windowY
}

var spritesAtCache [maxSpritesPerScanLine]oam

// spritesAt returns all sprites that are at the given X value, sorted by their
// drawing priority.  Sprites are loaded on a per-scan-line basis, so there's
// no need to check if sprites are at the current Y position.
func (vc *videoController) spritesAt(x uint8) []oam {
	spriteCount := 0

	for i := 0; i < vc.spriteCount; i++ {
		spriteX := vc.spritesOnScanLine[i].xPos

		// Check if the sprite this OAM entry corresponds to is in the given
		// point. Remember that a sprite's X and Y position is relative to the
		// bottom right of the sprite.
		if x < spriteX && int(x) >= int(spriteX)-spriteWidth {
			spritesAtCache[spriteCount] = vc.spritesOnScanLine[i]
			spriteCount++
		}
	}

	return spritesAtCache[:spriteCount]
}

// bgDotCode returns the dot code in the background layer at the given screen
// coordinates.
func (vc *videoController) bgDotCode(x, y uint8) uint8 {
	// Get the coordinates relative to the background and wrap them if
	// necessary
	bgX := int(x) + int(vc.scrollX)
	if bgX < 0 {
		bgX += bgWidth
	} else if bgX >= bgWidth {
		bgX -= bgWidth
	}
	bgY := int(y) + int(vc.scrollY)
	if bgY < 0 {
		bgY += bgHeight
	} else if bgY >= bgHeight {
		bgY -= bgHeight
	}

	// Get the tile this point is inside of
	tileOffset := (bgY/bgTileHeight)*bgWidthInTiles + (bgX / bgTileWidth)
	tileAddr := vc.lcdc.bgTileMapAddr + uint16(tileOffset)
	tile := vc.state.mmu.memory[tileAddr]

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
	tile := vc.state.mmu.memory[tileAddr]

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

	lower := vc.state.mmu.memory[tileDataAddr+uint16(inTileY*2)]
	upper := vc.state.mmu.memory[tileDataAddr+uint16((inTileY*2)+1)]

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

	lower := vc.state.mmu.memory[spriteDataAddr+uint16(inSpriteY*2)]
	upper := vc.state.mmu.memory[spriteDataAddr+uint16((inSpriteY*2)+1)]

	lower <<= uint(inSpriteX)
	upper <<= uint(inSpriteX)

	lowerBit := (lower & 0x80) >> 7
	upperBit := (upper & 0x80) >> 7
	return (upperBit << 1) | lowerBit
}

func (vc *videoController) destroy() {
	vc.driver.Close()
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
	vcMode1 = 1
	// The OAM RAM loading mode. OAM RAM may not be written to at this time.
	vcMode2 = 2
	// The VRAM and OAM RAM loading mode. VRAM and OAM RAM may not be written
	// to at this time.
	vcMode3 = 3
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
	lcdc := vc.state.mmu.memory[lcdcAddr]

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

// onLCDCWrite is called when the LCDC memory register is written to. It acts
// as a fast path for detecting LCD power toggles.
func (vc *videoController) onLCDCWrite(addr uint16, val uint8) uint8 {
	vc.lcdOn = val&0x80 == 0x80
	return val
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
	// The memory address where this OAM entry was retrieved from
	address       uint16
	yPos          uint8
	xPos          uint8
	spriteNumber  uint8
	priority      bool
	yFlip         bool
	xFlip         bool
	paletteNumber uint8
}

// loadSpritesOnScanLine loads all OAM entries from memory that are visible on
// the given scan line. These OAM entries are ordered by their priority,
// meaning that the first OAM entry that is at a given position should be drawn
// over all others at that position.
func (vc *videoController) loadSpritesOnScanLine(scanLine uint8) {
	vc.spriteCount = 0

	for i := 0; i < maxOAMEntries; i++ {
		entryStart := uint16(oamRAMAddr + (i * oamBytes))
		xPos := vc.state.mmu.memory[entryStart+1]
		if xPos == 0 {
			// This sprite is not on-screen
			continue
		}

		yPos := vc.state.mmu.memory[entryStart]
		var spriteTop int
		switch vc.lcdc.spriteSize {
		case spriteSize8x8:
			spriteTop = int(yPos) - spriteShortHeight
		case spriteSize8x16:
			spriteTop = int(yPos)
		}

		if int(scanLine) >= int(spriteTop) || int(scanLine) < int(yPos)-spriteTallHeight {
			// This sprite is not on this scan line
			continue
		}

		newOAM := oam{
			address:      entryStart,
			yPos:         yPos,
			xPos:         xPos,
			spriteNumber: vc.state.mmu.memory[entryStart+2],
		}

		flags := vc.state.mmu.memory[entryStart+3]
		newOAM.priority = flags&0x80 == 0x80
		newOAM.yFlip = flags&0x40 == 0x40
		newOAM.xFlip = flags&0x20 == 0x20

		if flags&0x10 == 0x10 {
			newOAM.paletteNumber = 1
		} else {
			newOAM.paletteNumber = 0
		}

		vc.spritesOnScanLine[vc.spriteCount] = newOAM
		vc.spriteCount++

		if vc.spriteCount == maxSpritesPerScanLine {
			// We've reached the maximum allowed sprites for this scan line. No
			// more can be drawn
			break
		}
	}

	// Sort sprites based on their drawing priority. Sprites with lower X
	// positions are drawn on top of sprites with higher X positions. If two
	// sprites are in the same X position, then the sprite with lower address
	// in OAM memory wins.
	sort.Slice(vc.spritesOnScanLine[:vc.spriteCount],
		// Return true if i has a higher priority than j.
		func(i, j int) bool {
			if vc.spritesOnScanLine[i].xPos == vc.spritesOnScanLine[j].xPos {
				return vc.spritesOnScanLine[i].address < vc.spritesOnScanLine[j].address
			} else if vc.spritesOnScanLine[i].xPos > vc.spritesOnScanLine[j].xPos {
				return false
			} else {
				return true
			}
		})
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
	stat := vc.state.mmu.memory[statAddr]

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
	statVal := vc.state.mmu.memory[statAddr]

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

	vc.state.mmu.memory[statAddr] = statVal
}

// loadPalette populates the given palette list with colors corresponding to
// the selections in the given palette data.
func loadPalette(palette *[4]color, paletteData uint8) {
	for dotData := 0; dotData < 0x04; dotData++ {
		paletteOption := paletteData & 0x03

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
		paletteData >>= 2
	}
}

// loadBGPalette inspects the BGP register value for palette information.
func (vc *videoController) loadBGPalette() {
	bgp := vc.state.mmu.memory[bgpAddr]

	loadPalette(&vc.bgPalette, bgp)
}

// loadSpritePalettes inspects the OBP0 or OPB1 register values and populates
// both sprite palettes.
func (vc *videoController) loadSpritePalettes() {
	opb0 := vc.state.mmu.memory[opb0Addr]
	opb1 := vc.state.mmu.memory[opb1Addr]

	loadPalette(&vc.spritePalette0, opb0)
	loadPalette(&vc.spritePalette1, opb1)
}

type color struct {
	r, g, b, a uint8
}
