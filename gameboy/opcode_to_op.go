package gameboy

import (
	"fmt"
)

// TODO(velovix): Document what instructions do to flags

// instruction is a function will run a CPU instruction operation and return
// the time in CPU cycles that the instruction took.
type instruction func(*State) int

// opcodeMapper is a lookup table that links opcodes to the corresponding
// instructions.
var opcodeMapper = []instruction{
	0x00: nop,
	0x01: ld16BitImm(regBC),
	0x02: ldToMem(regBC, regA),
	0x03: inc16Bit(regBC),
	0x04: inc8Bit(regB),
	0x05: dec8Bit(regB),
	0x06: ld8BitImm(regB),
	0x07: rlca,
	0x08: ldSPToMem,
	0x09: addToHL(regBC),
	0x0A: ldFromMem(regA, regBC),
	0x0B: dec16Bit(regBC),
	0x0C: inc8Bit(regC),
	0x0D: dec8Bit(regC),
	0x0E: ld8BitImm(regC),
	0x0F: rrca,
	0x10: stop,
	0x11: ld16BitImm(regDE),
	0x12: ldToMem(regDE, regA),
	0x13: inc16Bit(regDE),
	0x14: inc8Bit(regD),
	0x15: dec8Bit(regD),
	0x16: ld8BitImm(regD),
	0x17: rla,
	0x18: jr,
	0x19: addToHL(regDE),
	0x1A: ldFromMem(regA, regDE),
	0x1B: dec16Bit(regDE),
	0x1C: inc8Bit(regE),
	0x1D: dec8Bit(regE),
	0x1E: ld8BitImm(regE),
	0x1F: rra,
	0x20: jrIfFlag(zeroFlag, false),
	0x21: ld16BitImm(regHL),
	0x22: ldiToMem,
	0x23: inc16Bit(regHL),
	0x24: inc8Bit(regH),
	0x25: dec8Bit(regH),
	0x26: ld8BitImm(regH),
	0x27: daa,
	0x28: jrIfFlag(zeroFlag, true),
	0x29: addToHL(regHL),
	0x2A: ldiFromMem,
	0x2B: dec16Bit(regHL),
	0x2C: inc8Bit(regL),
	0x2D: dec8Bit(regL),
	0x2E: ld8BitImm(regL),
	0x2F: cpl,
	0x30: jrIfFlag(carryFlag, false),
	0x31: ld16BitImm(regSP),
	0x32: lddToMem,
	0x33: inc16Bit(regSP),
	0x34: incMemHL,
	0x35: decMemHL,
	0x36: ld8BitImmToMemHL,
	0x37: scf,
	0x38: jrIfFlag(carryFlag, true),
	0x39: addToHL(regSP),
	0x3A: lddFromMem,
	0x3B: dec16Bit(regSP),
	0x3C: inc8Bit(regA),
	0x3D: dec8Bit(regA),
	0x3E: ld8BitImm(regA),
	0x3F: ccf,
	0x40: ld(regB, regB),
	0x41: ld(regB, regC),
	0x42: ld(regB, regD),
	0x43: ld(regB, regE),
	0x44: ld(regB, regH),
	0x45: ld(regB, regL),
	0x46: ldFromMem(regB, regHL),
	0x47: ld(regB, regA),
	0x48: ld(regC, regB),
	0x49: ld(regC, regC),
	0x4A: ld(regC, regD),
	0x4B: ld(regC, regE),
	0x4C: ld(regC, regH),
	0x4D: ld(regC, regL),
	0x4E: ldFromMem(regC, regHL),
	0x4F: ld(regC, regA),
	0x50: ld(regD, regB),
	0x51: ld(regD, regC),
	0x52: ld(regD, regD),
	0x53: ld(regD, regE),
	0x54: ld(regD, regH),
	0x55: ld(regD, regL),
	0x56: ldFromMem(regD, regHL),
	0x57: ld(regD, regA),
	0x58: ld(regE, regB),
	0x59: ld(regE, regC),
	0x5A: ld(regE, regD),
	0x5B: ld(regE, regE),
	0x5C: ld(regE, regH),
	0x5D: ld(regE, regL),
	0x5E: ldFromMem(regE, regHL),
	0x5F: ld(regE, regA),
	0x60: ld(regH, regB),
	0x61: ld(regH, regC),
	0x62: ld(regH, regD),
	0x63: ld(regH, regE),
	0x64: ld(regH, regH),
	0x65: ld(regH, regL),
	0x66: ldFromMem(regH, regHL),
	0x67: ld(regH, regA),
	0x68: ld(regL, regB),
	0x69: ld(regL, regC),
	0x6A: ld(regL, regD),
	0x6B: ld(regL, regE),
	0x6C: ld(regL, regH),
	0x6D: ld(regL, regL),
	0x6E: ldFromMem(regL, regHL),
	0x6F: ld(regL, regA),
	0x70: ldToMem(regHL, regB),
	0x71: ldToMem(regHL, regC),
	0x72: ldToMem(regHL, regD),
	0x73: ldToMem(regHL, regE),
	0x74: ldToMem(regHL, regH),
	0x75: ldToMem(regHL, regL),
	0x76: halt,
	0x77: ldToMem(regHL, regA),
	0x78: ld(regA, regB),
	0x79: ld(regA, regC),
	0x7A: ld(regA, regD),
	0x7B: ld(regA, regE),
	0x7C: ld(regA, regH),
	0x7D: ld(regA, regL),
	0x7E: ldFromMem(regA, regHL),
	0x7F: ld(regA, regA),
	0x80: add(regB),
	0x81: add(regC),
	0x82: add(regD),
	0x83: add(regE),
	0x84: add(regH),
	0x85: add(regL),
	0x86: addFromMemHL,
	0x87: add(regA),
	0x88: adc(regB),
	0x89: adc(regC),
	0x8A: adc(regD),
	0x8B: adc(regE),
	0x8C: adc(regH),
	0x8D: adc(regL),
	0x8E: adcFromMemHL,
	0x8F: adc(regA),
	0x90: sub(regB),
	0x91: sub(regC),
	0x92: sub(regD),
	0x93: sub(regE),
	0x94: sub(regH),
	0x95: sub(regL),
	0x96: subFromMemHL,
	0x97: sub(regA),
	0x98: sbc(regB),
	0x99: sbc(regC),
	0x9A: sbc(regD),
	0x9B: sbc(regE),
	0x9C: sbc(regH),
	0x9D: sbc(regL),
	0x9E: sbcFromMemHL,
	0x9F: sbc(regA),
	0xA0: and(regB),
	0xA1: and(regC),
	0xA2: and(regD),
	0xA3: and(regE),
	0xA4: and(regH),
	0xA5: and(regL),
	0xA6: andFromMemHL,
	0xA7: and(regA),
	0xA8: xor(regB),
	0xA9: xor(regC),
	0xAA: xor(regD),
	0xAB: xor(regE),
	0xAC: xor(regH),
	0xAD: xor(regL),
	0xAE: xorFromMemHL,
	0xAF: xor(regA),
	0xB0: or(regB),
	0xB1: or(regC),
	0xB2: or(regD),
	0xB3: or(regE),
	0xB4: or(regH),
	0xB5: or(regL),
	0xB6: orFromMemHL,
	0xB7: or(regA),
	0xB8: cp(regB),
	0xB9: cp(regC),
	0xBA: cp(regD),
	0xBB: cp(regE),
	0xBC: cp(regH),
	0xBD: cp(regL),
	0xBE: cpFromMemHL,
	0xBF: cp(regA),
	0xC0: retIfFlag(zeroFlag, false),
	0xC1: pop(regBC),
	0xC2: jpIfFlag(zeroFlag, false),
	0xC3: jp,
	0xC4: callIfFlag(zeroFlag, false),
	0xC5: push(regBC),
	0xC6: add8BitImm,
	0xC7: rst(0x00),
	0xC8: retIfFlag(zeroFlag, true),
	0xC9: ret,
	0xCA: jpIfFlag(zeroFlag, true),
	0xCB: cbOpcodeDispatcher,
	0xCC: callIfFlag(zeroFlag, true),
	0xCD: call,
	0xCE: adc8BitImm,
	0xCF: rst(0x08),
	0xD0: retIfFlag(carryFlag, false),
	0xD1: pop(regDE),
	0xD2: jpIfFlag(carryFlag, false),
	0xD3: nil,
	0xD4: callIfFlag(carryFlag, false),
	0xD5: push(regDE),
	0xD6: sub8BitImm,
	0xD7: rst(0x10),
	0xD8: retIfFlag(carryFlag, true),
	0xD9: reti,
	0xDA: jpIfFlag(carryFlag, true),
	0xDB: nil,
	0xDC: callIfFlag(carryFlag, true),
	0xDD: nil,
	0xDE: sbc8BitImm,
	0xDF: rst(0x18),
	0xE0: ldhToMem,
	0xE1: pop(regHL),
	0xE2: ldToMemC,
	0xE3: nil,
	0xE4: nil,
	0xE5: push(regHL),
	0xE6: and8BitImm,
	0xE7: rst(0x20),
	0xE8: addToSP,
	0xE9: jpToHL,
	0xEA: ldTo16BitImmMem,
	0xEB: nil,
	0xEC: nil,
	0xED: nil,
	0xEE: xor8BitImm,
	0xEF: rst(0x28),
	0xF0: ldhFromMem,
	0xF1: pop(regAF),
	0xF2: ldFromMemC,
	0xF3: di,
	0xF4: nil,
	0xF5: push(regAF),
	0xF6: or8BitImm,
	0xF7: rst(0x30),
	0xF8: ldhl,
	0xF9: ldHLToSP,
	0xFA: ldFrom16BitImmMem,
	0xFB: ei,
	0xFC: nil,
	0xFD: nil,
	0xFE: cp8BitImm,
	0xFF: rst(0x38),
}

// runOpcode runs the operation that maps to the given opcode.
func runOpcode(state *State, opcode uint8) (cycles int, err error) {
	if printInstructions {
		fmt.Printf("@PC %#x | %#x: ", state.instructionStart, opcode)
	}

	instr := opcodeMapper[opcode]
	if instr == nil {
		return 0, fmt.Errorf("unknown opcode %#x", opcode)
	}
	return instr(state), nil
}

// runCBOpcode runs the operation that maps to the given opcode, assuming that
// it is CB-prefixed.
func runCBOpcode(state *State, opcode uint8) int {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	if printInstructions {
		fmt.Printf("%#x: ", opcode)
	}

	switch {
	// RLC r
	case opcode <= 0x05:
		return rlc(indexToRegister[opcode])(state)
	// RLC (HL)
	case opcode == 0x06:
		return rlcMemHL(state)
	// RLC A
	case opcode == 0x07:
		return rlc(regA)(state)
	// RRC r
	case opcode >= 0x08 && opcode <= 0x0D:
		regIndex := opcode - 0x08
		return rrc(indexToRegister[regIndex])(state)
	// RRC (HL)
	case opcode == 0x0E:
		return rrcMemHL(state)
	// RRC A
	case opcode == 0x0F:
		return rrc(regA)(state)
	// RL r
	case opcode >= 0x10 && opcode <= 0x15:
		regIndex := opcode - 0x10
		return rl(indexToRegister[regIndex])(state)
	// RL (HL)
	case opcode == 0x16:
		return rlMemHL(state)
	// RL A
	case opcode == 0x17:
		return rl(regA)(state)
	// RR A
	case opcode == 0x1F:
		return rr(regA)(state)
	// RR (HL)
	case opcode == 0x1E:
		return rrMemHL(state)
	// RR r
	case opcode >= 0x18 && opcode <= 0x1D:
		regIndex := opcode - 0x18
		return rr(indexToRegister[regIndex])(state)
	// SLA r
	case opcode >= 0x20 && opcode <= 0x25:
		regIndex := opcode - 0x20
		return sla(indexToRegister[regIndex])(state)
	// SLA (HL)
	case opcode == 0x26:
		return slaMemHL(state)
	// SLA A
	case opcode == 0x27:
		return sla(regA)(state)
	// SRA r
	case opcode >= 0x28 && opcode <= 0x2D:
		regIndex := opcode - 0x28
		return sra(indexToRegister[regIndex])(state)
	// SRA (HL)
	case opcode == 0x2E:
		return sraMemHL(state)
	// SRA A
	case opcode == 0x2F:
		return sra(regA)(state)
	// SWAP r
	case opcode >= 0x30 && opcode <= 0x35:
		return swap(indexToRegister[lowerNibble])(state)
	// SWAP (HL)
	case opcode == 0x36:
		return swapMemHL(state)
	// SWAP A
	case opcode == 0x37:
		return swap(regA)(state)
	// SRL a
	case opcode == 0x3F:
		return srl(regA)(state)
	// SRL (HL)
	case opcode == 0x3E:
		return srlMemHL(state)
	// SRL r
	case opcode >= 0x38 && opcode <= 0x3D:
		regIndex := opcode - 0x38
		return srl(indexToRegister[regIndex])(state)
	// BIT b,A
	case upperNibble >= 0x04 && upperNibble <= 0x07 &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bitNum := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0F {
			bitNum++
		}

		return bit(bitNum, regA)(state)
	// BIT b,(HL)
	case upperNibble >= 0x04 && upperNibble <= 0x07 &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return bitMemHL(bitNum)(state)
	// BIT b,r
	case upperNibble >= 0x04 && upperNibble <= 0x07 &&
		((lowerNibble >= 0x00 && lowerNibble <= 0x5) ||
			(lowerNibble >= 0x07 && lowerNibble <= 0x0D)):
		regIndex := lowerNibble
		if lowerNibble >= 0x08 {
			regIndex -= 0x08
		}

		bitNum := (upperNibble - 0x04) * 2
		if lowerNibble >= 0x08 {
			bitNum++
		}
		return bit(bitNum, indexToRegister[regIndex])(state)
	// RES b,A
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bit := (upperNibble - 0x08) * 2
		if lowerNibble == 0x0F {
			bit++
		}

		return res(bit, regA)(state)
	// RES b,(HL)
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x08) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return resMemHL(bitNum)(state)
	// RES b,r
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		((lowerNibble >= 0x00 && lowerNibble <= 0x5) ||
			(lowerNibble >= 0x07 && lowerNibble <= 0x0D)):
		regIndex := lowerNibble
		if lowerNibble >= 0x08 {
			regIndex -= 0x08
		}

		bit := (upperNibble - 0x08) * 2
		if lowerNibble >= 0x08 {
			bit++
		}
		return res(bit, indexToRegister[regIndex])(state)
	// SET b,A
	case upperNibble >= 0x0C && upperNibble <= 0x0F &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bit := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			bit++
		}

		return set(bit, regA)(state)
	// SET b,(HL)
	case upperNibble >= 0x0C && upperNibble <= 0x0F &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return setMemHL(bitNum)(state)
	// SET b,r
	case upperNibble >= 0x0C && upperNibble <= 0x0F &&
		((lowerNibble >= 0x00 && lowerNibble <= 0x5) ||
			(lowerNibble >= 0x07 && lowerNibble <= 0x0D)):
		regIndex := lowerNibble
		if lowerNibble >= 0x08 {
			regIndex -= 0x08
		}

		bit := (upperNibble - 0x0C) * 2
		if lowerNibble >= 0x08 {
			bit++
		}
		return set(bit, indexToRegister[regIndex])(state)
	default:
		// TODO(velovix): Handle this error properly
		panic(fmt.Sprintf("unknown CB opcode %v", opcode))
	}
}

// indexToRegister maps an index value to an 8-bit register. This helps reduce
// repetition for opcode sections that do the same thing on different
// registers, since the Game Boy is consistent about this ordering.
var indexToRegister = map[uint8]registerType{
	0: regB,
	1: regC,
	2: regD,
	3: regE,
	4: regH,
	5: regL,
}

// getConditionalStr creates a string representation of a conditional flag
// check.
func getConditionalStr(flagMask uint8, isSet bool) string {
	var conditional string
	switch flagMask {
	case zeroFlag:
		conditional = "Z"
	case carryFlag:
		conditional = "C"
	default:
		panic("unsupported JR conditional flag")
	}
	if !isSet {
		conditional = "N" + conditional
	}

	return conditional
}

// cbOpcodeDispatcher pops the next opcode from the program counter and
// dispatches the corresponding CB instruction.
func cbOpcodeDispatcher(state *State) int {
	cbOpcode := state.incrementPC()
	return runCBOpcode(state, cbOpcode)
}
