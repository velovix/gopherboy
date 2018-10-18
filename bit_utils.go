package main

// combine combines the given upper and lower uint8 values into a single
// uint16.
func combine(lower, upper uint8) uint16 {
	return (uint16(upper) << 8) | uint16(lower)
}

// split splits the given 8-bit value into its two 4-bit values, known
// unofficially as "nibbles".
func split(val uint8) (lower, upper uint8) {
	upper = uint8(val >> 4)
	lower = uint8(val & 0x0F)
	return lower, upper
}

// split16 splits the given 16-bit value into its two 8-bit values.
func split16(val uint16) (lower, upper uint8) {
	lower = uint8(val & 0xFF)
	upper = uint8(val >> 8)
	return lower, upper
}

// asSigned interprets the given uint8 as a two's complement and converts it to
// a signed int.
func asSigned(val uint8) int8 {
	// Check if the signed bit is set
	if val&0x80 == 0x80 {
		// Convert to a negative number
		signed := int8(^val)
		return -signed
	}

	// Value is positive, nothing to do
	return int8(val)
}

// isHalfCarry checks if a half carry would occur between two 8-bit integers if
// they were added.
//
// This algorithm extracts the first four bits of each register, adds them
// together, and checks the 5th bit to see if it's 1. If it is, that means the
// addition half-carried.
func isHalfCarry(a, b uint8) bool {
	return ((a&0xF)+(b&0xF))&0x10 == 0x10
}

// isCarry checks if there would be a carry past the 8th bit if two 8-bit
// integers were added.
func isCarry(a, b uint8) bool {
	return (uint16(a)+uint16(b))&0x100 == 0x100
}

// isHalfCarry16 checks if a half carry would occur between two 16-bit integers
// if they were added.
//
// This algorithm extracts the first 11 bits of each register, adds them
// together, and checks the 12th bit to see if it's 1. If it is, that means the
// addition half-carried.
func isHalfCarry16(a, b uint16) bool {
	return ((a&0x800)+(b&0x800))&0x1000 == 0x1000
}

// isCarry16 checks if there would be a carry past the 16th bit if two 16-bit
// integers were added.
func isCarry16(a, b uint16) bool {
	return (uint32(a)+uint32(b))&0x10000 == 0x10000
}
