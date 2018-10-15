package main

// combine combines the given upper and lower uint8 values into a single
// uint16.
func combine(lower, upper uint8) uint16 {
	return (uint16(upper) << 8) | uint16(lower)
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
