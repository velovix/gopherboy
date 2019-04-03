package gameboy

// VideoDriver describes an object can display RGBA frames.
type VideoDriver interface {
	// Render displays the given frame data on-screen.
	//
	// Each element of the frame data is 8-bit R, G, B, and A values laid out
	// in that order.
	Render(frameData []uint8) error
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

// InputDriver describes an object that can take in and report user input.
type InputDriver interface {
	// State returns the state of the given button since the last call to
	// Update. A true value indicates the button is pressed.
	State(Button) bool
	// Update updates the internal button state and returns true if a button
	// has been pressed since the last call.
	Update() bool
}

// SaveGameDriver describes an object that can save and load game saves.
type SaveGameDriver interface {
	// Save puts the game save in some persistent storage under a name.
	Save(name string, data []uint8) error
	// Load returns a game save from some persistent storage under the given
	// name.
	Load(name string) ([]uint8, error)
	// Has returns true if a game save exists under the given name.
	Has(name string) (bool, error)
}
