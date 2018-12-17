package main

// combine combines the given upper and lower nibbles into a single uint8.
func combine(lower, upper uint8) uint8 {
	return (upper << 4) | lower
}

// combine16 combines the given upper and lower uint8 values into a single
// uint16.
func combine16(lower, upper uint8) uint16 {
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

// isHalfCarry checks if a half carry would occur between two or more 8-bit
// integers if they were added.
//
// This algorithm extracts the first four bits of each integer, adds them
// together, and checks the 5th bit to see if it's 1. If it is, that means the
// addition half-carried.
func isHalfCarry(vals ...uint8) bool {
	sumSoFar := vals[0]
	for i := 1; i < len(vals); i++ {
		if ((sumSoFar&0xF)+(vals[i]&0xF))&0x10 == 0x10 {
			return true
		}
		sumSoFar += vals[i]
	}

	return false
}

// isBorrow checks if a borrow would occur between two or more 8-bit integers
// if they are subtracted. In this case, a borrow is equivalent to an
// underflow.
//
// This algorithm is simple. If the number we're subtracting by is larger than
// the original number, a borrow must be necessary
func isBorrow(vals ...uint8) bool {
	diffSoFar := vals[0]
	for i := 1; i < len(vals); i++ {
		if diffSoFar < vals[i] {
			return true
		}
		diffSoFar -= vals[i]
	}

	return false
}

// isHalfBorrow checks if a half borrow would occur between two 8-bit integers
// if they were subtracted.
//
// This algorithm extracts the first four bits of each integer and checks if
// the a value bits are less than the b value bits. This tells us if a borrow
// will be necessary.
func isHalfBorrow(vals ...uint8) bool {
	diffSoFar := vals[0]
	for i := 1; i < len(vals); i++ {
		if diffSoFar&0xF < vals[i]&0xF {
			return true
		}
		diffSoFar -= vals[i]
	}

	return false
}

// isCarry checks if there would be a carry past the 8th bit if two or more
// 8-bit integers were added.
func isCarry(vals ...uint8) bool {
	sumSoFar := vals[0]
	for i := 1; i < len(vals); i++ {
		if (uint16(sumSoFar)+uint16(vals[i]))&0x100 == 0x100 {
			return true
		}
		sumSoFar += vals[i]
	}

	return false
}

// isHalfCarry16 checks if a half carry would occur between two 16-bit integers
// if they were added.
//
// This algorithm extracts the first 11 bits of each register, adds them
// together, and checks the 12th bit to see if it's 1. If it is, that means the
// addition half-carried.
func isHalfCarry16(a, b uint16) bool {
	return ((a&0xFFF)+(b&0xFFF))&0x1000 == 0x1000
}

// isCarry16 checks if there would be a carry past the 16th bit if two 16-bit
// integers were added.
func isCarry16(a, b uint16) bool {
	return (uint32(a)+uint32(b))&0x10000 == 0x10000
}
