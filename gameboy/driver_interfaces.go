package gameboy

type VideoDriver interface {
	// Render displays the given frame data on-screen.
	//
	// Each element of the frame data is a combination of 8-bit R, G, B, and A
	// values in that order.
	Render(frameData []uint32) error
	// Clear clears the display in preparation for another frame.
	Clear()
	// Close de-initializes the driver.
	Close()
}

type Button int

const (
	_ Button = iota
	ButtonStart
	ButtonSelect
	ButtonB
	ButtonA
	ButtonDown
	ButtonUp
	ButtonLeft
	ButtonRight
)

type InputDriver interface {
	// State returns the state of the given button since the last call to
	// Update. A true value indicates the button is pressed.
	State(Button) bool
	// Update updates the internal button state and returns true if a button
	// has been pressed since the last call.
	Update() bool
}
