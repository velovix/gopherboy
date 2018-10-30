package main

import (
	"fmt"
)

// TODO(velovix): Document what instructions do to flags

// runOpcode runs the operation that maps to the given opcode.
func runOpcode(env *environment, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	if printInstructions {
		fmt.Printf("%#x: ", opcode)
	}

	switch {
	// NOP
	case opcode == 0x00:
		return nop(env), nil
	// The CB prefix. Comes before an extended instruction set.
	case opcode == 0xCB:
		cbOpcode := env.incrementPC()
		return runCBOpcode(env, cbOpcode)
	// STOP 0
	case opcode == 0x10:
		panic("STOP 0 is not supported")
	// DI
	case opcode == 0xF3:
		return di(env), nil
	// EI
	case opcode == 0xFB:
		return ei(env), nil
	// HALT
	case opcode == 0x76:
		return halt(env), nil
	// LD (BC),A
	case opcode == 0x02:
		return ldToMem(env, regBC, regA), nil
	// LD (DE),A
	case opcode == 0x12:
		return ldToMem(env, regDE, regA), nil
	// LD A,(BC)
	case opcode == 0x0A:
		return ldFromMem(env, regA, regBC), nil
	// LD A,(DE)
	case opcode == 0x1A:
		return ldFromMem(env, regA, regDE), nil
	// LD (a16),A
	case opcode == 0xEA:
		return ldTo16BitImmMem(env), nil
	// LD A,(a16)
	case opcode == 0xFA:
		return ldFrom16BitImmMem(env), nil
	// LD (C),A
	case opcode == 0xE2:
		return ldToMemC(env), nil
	// LD A,(C)
	case opcode == 0xF2:
		return ldFromMemC(env), nil
	// LDH (a8),A
	case opcode == 0xE0:
		return ldhToMem(env), nil
	// LDH A,(a8)
	case opcode == 0xF0:
		return ldhFromMem(env), nil
	// LD HL,SP+r8
	case opcode == 0xF8:
		return ldhl(env), nil
	// LD SP,HL
	case opcode == 0xF9:
		return ldHLToSP(env), nil
	// JR a8
	case opcode == 0x18:
		return jr(env), nil
	// JR NZ,a8
	case opcode == 0x20:
		return jrIfFlag(env, zeroFlag, false), nil
	// JR Z,a8
	case opcode == 0x28:
		return jrIfFlag(env, zeroFlag, true), nil
	// JR NC,a8
	case opcode == 0x30:
		return jrIfFlag(env, carryFlag, false), nil
	// JR C,a8
	case opcode == 0x38:
		return jrIfFlag(env, carryFlag, true), nil
	// JP a16
	case opcode == 0xC3:
		return jp(env), nil
	// JP NZ,a16
	case opcode == 0xC2:
		return jpIfFlag(env, zeroFlag, false), nil
	// JP Z,a16
	case opcode == 0xCA:
		return jpIfFlag(env, zeroFlag, true), nil
	// JP NC,a16
	case opcode == 0xD2:
		return jpIfFlag(env, carryFlag, false), nil
	// JP C,a16
	case opcode == 0xDA:
		return jpIfFlag(env, carryFlag, true), nil
	// JP (HL)
	case opcode == 0xE9:
		return jpToHL(env), nil
	// RET
	case opcode == 0xC9:
		return ret(env), nil
	// RET NZ
	case opcode == 0xC0:
		return retIfFlag(env, zeroFlag, false), nil
	// RET Z
	case opcode == 0xC8:
		return retIfFlag(env, zeroFlag, true), nil
	// RET NC
	case opcode == 0xD0:
		return retIfFlag(env, carryFlag, false), nil
	// RET C
	case opcode == 0xD8:
		return retIfFlag(env, carryFlag, true), nil
	// RETI
	case opcode == 0xD9:
		return reti(env), nil
	// CALL a16
	case opcode == 0xCD:
		return call(env), nil
	// CALL NZ,a16
	case opcode == 0xC4:
		return callIfFlag(env, zeroFlag, false), nil
	// CALL Z,a16
	case opcode == 0xCC:
		return callIfFlag(env, zeroFlag, true), nil
	// CALL NC,a16
	case opcode == 0xD4:
		return callIfFlag(env, carryFlag, false), nil
	// CALL C,a16
	case opcode == 0xDC:
		return callIfFlag(env, carryFlag, true), nil
	// RLCA
	case opcode == 0x07:
		return rlca(env), nil
	// RLA
	case opcode == 0x17:
		return rla(env), nil
	// RRCA
	case opcode == 0x0F:
		return rrca(env), nil
	// RRA
	case opcode == 0x1F:
		return rra(env), nil
	// LD A,(HL+)
	case opcode == 0x2A:
		return ldiFromMem(env), nil
	// LD A,(HL-)
	case opcode == 0x3A:
		return lddFromMem(env), nil
	// LD (HL+),A
	case opcode == 0x22:
		return ldiToMem(env), nil
	// LD (HL-),A
	case opcode == 0x32:
		return lddToMem(env), nil
	// ADD HL,nn
	case lowerNibble == 0x09 && upperNibble <= 0x03:
		return addToHL(env, indexTo16BitRegister[upperNibble]), nil
	// LD nn,SP
	case opcode == 0x08:
		return ldSPToMem(env), nil
	// LD r,d16
	case lowerNibble == 0x01 && upperNibble <= 0x03:
		return ld16BitImm(env, indexTo16BitRegister[upperNibble]), nil
	// LD A,(HL)
	case opcode == 0x7E:
		return ldFromMem(env, regA, regHL), nil
	// LD r,(HL)
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ldFromMem(env, indexToRegister[regIndex], regHL), nil
	// LD (HL),A
	case opcode == 0x77:
		return ldToMem(env, regHL, regA), nil
	// LD A,A
	case opcode == 0x7F:
		return ld(env, regA, regA), nil
	// LD r,A
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0F {
			regIndex++
		}
		return ld(env, indexToRegister[regIndex], regA), nil
	// LD B,r
	case upperNibble == 0x04 && lowerNibble <= 0x05:
		return ld(env, regB, indexToRegister[lowerNibble]), nil
	// LD C,r
	case upperNibble == 0x04 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regC, indexToRegister[regIndex]), nil
	// LD D,r
	case upperNibble == 0x05 && lowerNibble <= 0x05:
		return ld(env, regD, indexToRegister[lowerNibble]), nil
	// LD E,r
	case upperNibble == 0x05 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regE, indexToRegister[regIndex]), nil
	// LD H,r
	case upperNibble == 0x06 && lowerNibble <= 0x05:
		return ld(env, regH, indexToRegister[lowerNibble]), nil
	// LD L,r
	case upperNibble == 0x06 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regL, indexToRegister[regIndex]), nil
	// LD (HL),r
	case upperNibble == 0x07 && lowerNibble <= 0x05:
		return ldToMem(env, regHL, indexToRegister[lowerNibble]), nil
	// LD A,r
	case upperNibble == 0x07 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(env, regA, indexToRegister[regIndex]), nil
	// LD (HL),d8
	case opcode == 0x36:
		return ld8BitImmToMemHL(env), nil
	// LD A,d8
	case opcode == 0x3E:
		return ld8BitImm(env, regA), nil
	// LD r,d8
	case upperNibble <= 0x02 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ld8BitImm(env, indexToRegister[regIndex]), nil
	// ADD A,A
	case opcode == 0x87:
		return add(env, regA), nil
	// ADD A,(HL)
	case opcode == 0x86:
		return addFromMemHL(env), nil
	// ADD A,r
	case opcode >= 0x80 && opcode <= 0x85:
		return add(env, indexToRegister[lowerNibble]), nil
	// ADD A,d8
	case opcode == 0xC6:
		return add8BitImm(env), nil
	// ADD SP,n
	case opcode == 0xE8:
		return addToSP(env), nil
	// ADC A
	case opcode == 0x8F:
		return adc(env, regA), nil
	// ADC (HL)
	case opcode == 0x8E:
		return adcFromMemHL(env), nil
	// ADC r
	case opcode >= 0x88 && opcode <= 0x8D:
		regIndex := opcode - 0x88
		return adc(env, indexToRegister[regIndex]), nil
	// ADC n
	case opcode == 0xCE:
		return adc8BitImm(env), nil
	// SUB A
	case opcode == 0x97:
		return sub(env, regA), nil
	// SUB (HL)
	case opcode == 0x96:
		return subFromMemHL(env), nil
	// SUB r
	case opcode >= 0x90 && opcode <= 0x95:
		return sub(env, indexToRegister[lowerNibble]), nil
	// SUB n
	case opcode == 0xD6:
		return sub8BitImm(env), nil
	// SBC A,A
	case opcode == 0x9F:
		return sbc(env, regA), nil
	// SBC A,(HL)
	case opcode == 0x9E:
		return sbcFromMemHL(env), nil
	// SBC A,d8
	case opcode == 0xDE:
		return sbc8BitImm(env), nil
	// SBC r
	case opcode >= 0x98 && opcode <= 0x9D:
		regIndex := opcode - 0x98
		return sbc(env, indexToRegister[regIndex]), nil
	// AND A
	case opcode == 0xA7:
		return and(env, regA), nil
	// AND (HL)
	case opcode == 0xA6:
		return andFromMemHL(env), nil
	// AND r
	case opcode >= 0xA0 && opcode <= 0xA5:
		return and(env, indexToRegister[lowerNibble]), nil
	// AND n
	case opcode == 0xE6:
		return and8BitImm(env), nil
	// OR A
	case opcode == 0xB7:
		return or(env, regA), nil
	// OR (HL)
	case opcode == 0xB6:
		return orFromMemHL(env), nil
	// OR r
	case opcode >= 0xB0 && opcode <= 0xB5:
		return or(env, indexToRegister[lowerNibble]), nil
	// OR n
	case opcode == 0xF6:
		return or8BitImm(env), nil
	// XOR A
	case opcode == 0xAF:
		return xor(env, regA), nil
	// XOR (HL)
	case opcode == 0xAE:
		return xorFromMemHL(env), nil
	// XOR r
	case opcode >= 0xA8 && opcode <= 0xAD:
		regIndex := lowerNibble - 0x08
		return xor(env, indexToRegister[regIndex]), nil
	// XOR n
	case opcode == 0xEE:
		return xor8BitImm(env), nil
	// CP A
	case opcode == 0xBF:
		return cp(env, regA), nil
	// CP (HL)
	case opcode == 0xBE:
		return cpFromMemHL(env), nil
	// CP n
	case opcode == 0xFE:
		return cp8BitImm(env), nil
	// CP r
	case opcode >= 0xB8 && opcode <= 0xBD:
		regIndex := lowerNibble - 0x08
		return cp(env, indexToRegister[regIndex]), nil
	// INC (HL)
	case opcode == 0x34:
		return incMemHL(env), nil
	// INC A
	case opcode == 0x3C:
		return inc8Bit(env, regA), nil
	// INC r
	case upperNibble <= 0x02 && (lowerNibble == 0x04 || lowerNibble == 0x0C):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0C {
			regIndex++
		}

		return inc8Bit(env, indexToRegister[regIndex]), nil
	// INC ss
	case upperNibble <= 0x03 && lowerNibble == 0x03:
		return inc16Bit(env, indexTo16BitRegister[upperNibble]), nil
	// DEC (HL)
	case opcode == 0x35:
		return decMemHL(env), nil
	// DEC A
	case opcode == 0x3D:
		return dec8Bit(env, regA), nil
	// DEC r
	case upperNibble <= 0x02 && (lowerNibble == 0x05 || lowerNibble == 0x0D):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0D {
			regIndex += 1
		}

		return dec8Bit(env, indexToRegister[regIndex]), nil
	// DEC ss
	case upperNibble <= 0x03 && lowerNibble == 0x0B:
		return dec16Bit(env, indexTo16BitRegister[upperNibble]), nil
	// POP ss
	case upperNibble >= 0x0C && upperNibble <= 0x0E && lowerNibble == 0x01:
		regIndex := upperNibble - 0x0C
		return pop(env, indexTo16BitRegister[regIndex]), nil
	// POP AF
	case opcode == 0xF1:
		return pop(env, regAF), nil
	// PUSH ss
	case upperNibble >= 0x0C && upperNibble <= 0x0E && lowerNibble == 0x05:
		regIndex := upperNibble - 0x0C
		return push(env, indexTo16BitRegister[regIndex]), nil
	// PUSH AF
	case opcode == 0xF5:
		return push(env, regAF), nil
	// RST address
	case upperNibble >= 0x0C && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		restartIndex := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			restartIndex++
		}

		return rst(env, indexToRestartAddress[restartIndex]), nil
	// CPL
	case opcode == 0x2F:
		return cpl(env), nil
	// CCF
	case opcode == 0x3F:
		return ccf(env), nil
	// SCF
	case opcode == 0x37:
		return scf(env), nil
	default:
		return 0, fmt.Errorf("unknown opcode %#x", opcode)
	}
}

// runCBOpcode runs the operation that maps to the given opcode, assuming that
// it is CB-prefixed.
func runCBOpcode(env *environment, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	if printInstructions {
		fmt.Printf("%#x: ", opcode)
	}

	switch {
	// RR A
	case opcode == 0x1F:
		return rr(env, regA), nil
	// RR (HL)
	case opcode == 0x1E:
		return rrMemHL(env), nil
	// RR r
	case opcode >= 0x18 && opcode <= 0x1D:
		regIndex := opcode - 0x18
		return rr(env, indexToRegister[regIndex]), nil
	// SRL a
	case opcode == 0x3F:
		return srl(env, regA), nil
	// SRL (HL)
	case opcode == 0x3E:
		return srlMemHL(env), nil
	// SRL r
	case opcode >= 0x38 && opcode <= 0x3D:
		regIndex := opcode - 0x38
		return srl(env, indexToRegister[regIndex]), nil
	// RES b,A
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		lowerNibble == 0x07 || lowerNibble == 0x0F:
		bit := (upperNibble - 0x08) * 2
		if lowerNibble == 0x0F {
			bit++
		}

		return res(env, bit, regA), nil
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
		return res(env, bit, indexToRegister[regIndex]), nil
	// SWAP A
	case opcode == 0x37:
		return swap(env, regA), nil
	// SWAP r
	case opcode >= 0x30 && opcode <= 0x35:
		return swap(env, indexToRegister[lowerNibble]), nil
	default:
		return 0, fmt.Errorf("unknown CB-prefixed opcode %#x", opcode)
	}
}

// indexToRegister maps an index value to an 8-bit register. This helps reduce
// repetition for opcode sections that do the same thing on different
// registers, since the Gameboy is consistent about this ordering.
var indexToRegister = map[uint8]registerType{
	0: regB,
	1: regC,
	2: regD,
	3: regE,
	4: regH,
	5: regL,
}

// indexTo16BitRegister maps an index value to a 16-bit register. This serves
// the same purpose as indexToRegister, but for 16-bit registers.
var indexTo16BitRegister = map[uint8]registerType{
	0: regBC,
	1: regDE,
	2: regHL,
	3: regSP,
}

// indexToRestartAddress maps an index value to a place in memory. These values
// are specific to the RST instruction. This helps reduce repetition when
// mapping the different flavors of that instruction from their opcodes.
var indexToRestartAddress = map[uint8]uint16{
	0: 0x0000,
	1: 0x0008,
	2: 0x0010,
	3: 0x0018,
	4: 0x0020,
	5: 0x0028,
	6: 0x0030,
	7: 0x0038,
}

// getConditionalStr creates a string representation of a conditional flag
// check.
func getConditionalStr(flagMask uint16, isSet bool) string {
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
