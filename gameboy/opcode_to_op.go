package gameboy

import (
	"fmt"
)

// TODO(velovix): Document what instructions do to flags

// runOpcode runs the operation that maps to the given opcode.
func runOpcode(state *State, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	if printInstructions {
		fmt.Printf("@PC %#x | %#x: ", state.instructionStart, opcode)
	}

	switch {
	// NOP
	case opcode == 0x00:
		return nop(state), nil
	// The CB prefix. Comes before an extended instruction set.
	case opcode == 0xCB:
		cbOpcode := state.incrementPC()
		return runCBOpcode(state, cbOpcode)
	// STOP 0
	case opcode == 0x10:
		return stop(state), nil
	// DI
	case opcode == 0xF3:
		return di(state), nil
	// EI
	case opcode == 0xFB:
		return ei(state), nil
	// HALT
	case opcode == 0x76:
		return halt(state), nil
	// LD (BC),A
	case opcode == 0x02:
		return ldToMem(state, regBC, regA), nil
	// LD (DE),A
	case opcode == 0x12:
		return ldToMem(state, regDE, regA), nil
	// LD A,(BC)
	case opcode == 0x0A:
		return ldFromMem(state, regA, regBC), nil
	// LD A,(DE)
	case opcode == 0x1A:
		return ldFromMem(state, regA, regDE), nil
	// LD (a16),A
	case opcode == 0xEA:
		return ldTo16BitImmMem(state), nil
	// LD A,(a16)
	case opcode == 0xFA:
		return ldFrom16BitImmMem(state), nil
	// LD (C),A
	case opcode == 0xE2:
		return ldToMemC(state), nil
	// LD A,(C)
	case opcode == 0xF2:
		return ldFromMemC(state), nil
	// LDH (a8),A
	case opcode == 0xE0:
		return ldhToMem(state), nil
	// LDH A,(a8)
	case opcode == 0xF0:
		return ldhFromMem(state), nil
	// LD HL,SP+r8
	case opcode == 0xF8:
		return ldhl(state), nil
	// LD SP,HL
	case opcode == 0xF9:
		return ldHLToSP(state), nil
	// JR a8
	case opcode == 0x18:
		return jr(state), nil
	// JR NZ,a8
	case opcode == 0x20:
		return jrIfFlag(state, zeroFlag, false), nil
	// JR Z,a8
	case opcode == 0x28:
		return jrIfFlag(state, zeroFlag, true), nil
	// JR NC,a8
	case opcode == 0x30:
		return jrIfFlag(state, carryFlag, false), nil
	// JR C,a8
	case opcode == 0x38:
		return jrIfFlag(state, carryFlag, true), nil
	// JP a16
	case opcode == 0xC3:
		return jp(state), nil
	// JP NZ,a16
	case opcode == 0xC2:
		return jpIfFlag(state, zeroFlag, false), nil
	// JP Z,a16
	case opcode == 0xCA:
		return jpIfFlag(state, zeroFlag, true), nil
	// JP NC,a16
	case opcode == 0xD2:
		return jpIfFlag(state, carryFlag, false), nil
	// JP C,a16
	case opcode == 0xDA:
		return jpIfFlag(state, carryFlag, true), nil
	// JP (HL)
	case opcode == 0xE9:
		return jpToHL(state), nil
	// RET
	case opcode == 0xC9:
		return ret(state), nil
	// RET NZ
	case opcode == 0xC0:
		return retIfFlag(state, zeroFlag, false), nil
	// RET Z
	case opcode == 0xC8:
		return retIfFlag(state, zeroFlag, true), nil
	// RET NC
	case opcode == 0xD0:
		return retIfFlag(state, carryFlag, false), nil
	// RET C
	case opcode == 0xD8:
		return retIfFlag(state, carryFlag, true), nil
	// RETI
	case opcode == 0xD9:
		return reti(state), nil
	// CALL a16
	case opcode == 0xCD:
		return call(state), nil
	// CALL NZ,a16
	case opcode == 0xC4:
		return callIfFlag(state, zeroFlag, false), nil
	// CALL Z,a16
	case opcode == 0xCC:
		return callIfFlag(state, zeroFlag, true), nil
	// CALL NC,a16
	case opcode == 0xD4:
		return callIfFlag(state, carryFlag, false), nil
	// CALL C,a16
	case opcode == 0xDC:
		return callIfFlag(state, carryFlag, true), nil
	// RLCA
	case opcode == 0x07:
		return rlca(state), nil
	// RLA
	case opcode == 0x17:
		return rla(state), nil
	// RRCA
	case opcode == 0x0F:
		return rrca(state), nil
	// RRA
	case opcode == 0x1F:
		return rra(state), nil
	// LD A,(HL+)
	case opcode == 0x2A:
		return ldiFromMem(state), nil
	// LD A,(HL-)
	case opcode == 0x3A:
		return lddFromMem(state), nil
	// LD (HL+),A
	case opcode == 0x22:
		return ldiToMem(state), nil
	// LD (HL-),A
	case opcode == 0x32:
		return lddToMem(state), nil
	// ADD HL,nn
	case lowerNibble == 0x09 && upperNibble <= 0x03:
		return addToHL(state, indexTo16BitRegister[upperNibble]), nil
	// LD nn,SP
	case opcode == 0x08:
		return ldSPToMem(state), nil
	// LD r,d16
	case lowerNibble == 0x01 && upperNibble <= 0x03:
		return ld16BitImm(state, indexTo16BitRegister[upperNibble]), nil
	// LD A,(HL)
	case opcode == 0x7E:
		return ldFromMem(state, regA, regHL), nil
	// LD r,(HL)
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ldFromMem(state, indexToRegister[regIndex], regHL), nil
	// LD (HL),A
	case opcode == 0x77:
		return ldToMem(state, regHL, regA), nil
	// LD A,A
	case opcode == 0x7F:
		return ld(state, regA, regA), nil
	// LD r,A
	case upperNibble >= 0x04 && upperNibble <= 0x06 && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		regIndex := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0F {
			regIndex++
		}
		return ld(state, indexToRegister[regIndex], regA), nil
	// LD B,r
	case upperNibble == 0x04 && lowerNibble <= 0x05:
		return ld(state, regB, indexToRegister[lowerNibble]), nil
	// LD C,r
	case upperNibble == 0x04 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(state, regC, indexToRegister[regIndex]), nil
	// LD D,r
	case upperNibble == 0x05 && lowerNibble <= 0x05:
		return ld(state, regD, indexToRegister[lowerNibble]), nil
	// LD E,r
	case upperNibble == 0x05 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(state, regE, indexToRegister[regIndex]), nil
	// LD H,r
	case upperNibble == 0x06 && lowerNibble <= 0x05:
		return ld(state, regH, indexToRegister[lowerNibble]), nil
	// LD L,r
	case upperNibble == 0x06 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(state, regL, indexToRegister[regIndex]), nil
	// LD (HL),r
	case upperNibble == 0x07 && lowerNibble <= 0x05:
		return ldToMem(state, regHL, indexToRegister[lowerNibble]), nil
	// LD A,r
	case upperNibble == 0x07 && lowerNibble >= 0x08 && lowerNibble <= 0x0D:
		regIndex := lowerNibble - 0x08
		return ld(state, regA, indexToRegister[regIndex]), nil
	// LD (HL),d8
	case opcode == 0x36:
		return ld8BitImmToMemHL(state), nil
	// LD A,d8
	case opcode == 0x3E:
		return ld8BitImm(state, regA), nil
	// LD r,d8
	case upperNibble <= 0x02 && (lowerNibble == 0x06 || lowerNibble == 0x0E):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0E {
			regIndex++
		}
		return ld8BitImm(state, indexToRegister[regIndex]), nil
	// ADD A,A
	case opcode == 0x87:
		return add(state, regA), nil
	// ADD A,(HL)
	case opcode == 0x86:
		return addFromMemHL(state), nil
	// ADD A,r
	case opcode >= 0x80 && opcode <= 0x85:
		return add(state, indexToRegister[lowerNibble]), nil
	// ADD A,d8
	case opcode == 0xC6:
		return add8BitImm(state), nil
	// ADD SP,n
	case opcode == 0xE8:
		return addToSP(state), nil
	// ADC A
	case opcode == 0x8F:
		return adc(state, regA), nil
	// ADC (HL)
	case opcode == 0x8E:
		return adcFromMemHL(state), nil
	// ADC r
	case opcode >= 0x88 && opcode <= 0x8D:
		regIndex := opcode - 0x88
		return adc(state, indexToRegister[regIndex]), nil
	// ADC n
	case opcode == 0xCE:
		return adc8BitImm(state), nil
	// SUB A
	case opcode == 0x97:
		return sub(state, regA), nil
	// SUB (HL)
	case opcode == 0x96:
		return subFromMemHL(state), nil
	// SUB r
	case opcode >= 0x90 && opcode <= 0x95:
		return sub(state, indexToRegister[lowerNibble]), nil
	// SUB n
	case opcode == 0xD6:
		return sub8BitImm(state), nil
	// SBC A,A
	case opcode == 0x9F:
		return sbc(state, regA), nil
	// SBC A,(HL)
	case opcode == 0x9E:
		return sbcFromMemHL(state), nil
	// SBC A,d8
	case opcode == 0xDE:
		return sbc8BitImm(state), nil
	// SBC r
	case opcode >= 0x98 && opcode <= 0x9D:
		regIndex := opcode - 0x98
		return sbc(state, indexToRegister[regIndex]), nil
	// AND A
	case opcode == 0xA7:
		return and(state, regA), nil
	// AND (HL)
	case opcode == 0xA6:
		return andFromMemHL(state), nil
	// AND r
	case opcode >= 0xA0 && opcode <= 0xA5:
		return and(state, indexToRegister[lowerNibble]), nil
	// AND n
	case opcode == 0xE6:
		return and8BitImm(state), nil
	// OR A
	case opcode == 0xB7:
		return or(state, regA), nil
	// OR (HL)
	case opcode == 0xB6:
		return orFromMemHL(state), nil
	// OR r
	case opcode >= 0xB0 && opcode <= 0xB5:
		return or(state, indexToRegister[lowerNibble]), nil
	// OR n
	case opcode == 0xF6:
		return or8BitImm(state), nil
	// XOR A
	case opcode == 0xAF:
		return xor(state, regA), nil
	// XOR (HL)
	case opcode == 0xAE:
		return xorFromMemHL(state), nil
	// XOR r
	case opcode >= 0xA8 && opcode <= 0xAD:
		regIndex := lowerNibble - 0x08
		return xor(state, indexToRegister[regIndex]), nil
	// XOR n
	case opcode == 0xEE:
		return xor8BitImm(state), nil
	// CP A
	case opcode == 0xBF:
		return cp(state, regA), nil
	// CP (HL)
	case opcode == 0xBE:
		return cpFromMemHL(state), nil
	// CP n
	case opcode == 0xFE:
		return cp8BitImm(state), nil
	// CP r
	case opcode >= 0xB8 && opcode <= 0xBD:
		regIndex := lowerNibble - 0x08
		return cp(state, indexToRegister[regIndex]), nil
	// INC (HL)
	case opcode == 0x34:
		return incMemHL(state), nil
	// INC A
	case opcode == 0x3C:
		return inc8Bit(state, regA), nil
	// INC r
	case upperNibble <= 0x02 && (lowerNibble == 0x04 || lowerNibble == 0x0C):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0C {
			regIndex++
		}

		return inc8Bit(state, indexToRegister[regIndex]), nil
	// INC ss
	case upperNibble <= 0x03 && lowerNibble == 0x03:
		return inc16Bit(state, indexTo16BitRegister[upperNibble]), nil
	// DEC (HL)
	case opcode == 0x35:
		return decMemHL(state), nil
	// DEC A
	case opcode == 0x3D:
		return dec8Bit(state, regA), nil
	// DEC r
	case upperNibble <= 0x02 && (lowerNibble == 0x05 || lowerNibble == 0x0D):
		regIndex := upperNibble * 2
		if lowerNibble == 0x0D {
			regIndex++
		}

		return dec8Bit(state, indexToRegister[regIndex]), nil
	// DEC ss
	case upperNibble <= 0x03 && lowerNibble == 0x0B:
		return dec16Bit(state, indexTo16BitRegister[upperNibble]), nil
	// POP ss
	case upperNibble >= 0x0C && upperNibble <= 0x0E && lowerNibble == 0x01:
		regIndex := upperNibble - 0x0C
		return pop(state, indexTo16BitRegister[regIndex]), nil
	// POP AF
	case opcode == 0xF1:
		return pop(state, regAF), nil
	// PUSH ss
	case upperNibble >= 0x0C && upperNibble <= 0x0E && lowerNibble == 0x05:
		regIndex := upperNibble - 0x0C
		return push(state, indexTo16BitRegister[regIndex]), nil
	// PUSH AF
	case opcode == 0xF5:
		return push(state, regAF), nil
	// RST address
	case upperNibble >= 0x0C && (lowerNibble == 0x07 || lowerNibble == 0x0F):
		restartIndex := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			restartIndex++
		}

		return rst(state, indexToRestartAddress[restartIndex]), nil
	// CPL
	case opcode == 0x2F:
		return cpl(state), nil
	// CCF
	case opcode == 0x3F:
		return ccf(state), nil
	// SCF
	case opcode == 0x37:
		return scf(state), nil
	// DAA
	case opcode == 0x27:
		return daa(state), nil
	default:
		return 0, fmt.Errorf("unknown opcode %#x", opcode)
	}
}

// runCBOpcode runs the operation that maps to the given opcode, assuming that
// it is CB-prefixed.
func runCBOpcode(state *State, opcode uint8) (cycles int, err error) {
	// Splits the 8-bit opcode into two nibbles
	lowerNibble, upperNibble := split(opcode)

	if printInstructions {
		fmt.Printf("%#x: ", opcode)
	}

	switch {
	// RLC r
	case opcode <= 0x05:
		return rlc(state, indexToRegister[opcode]), nil
	// RLC (HL)
	case opcode == 0x06:
		return rlcMemHL(state), nil
	// RLC A
	case opcode == 0x07:
		return rlc(state, regA), nil
	// RRC r
	case opcode >= 0x08 && opcode <= 0x0D:
		regIndex := opcode - 0x08
		return rrc(state, indexToRegister[regIndex]), nil
	// RRC (HL)
	case opcode == 0x0E:
		return rrcMemHL(state), nil
	// RRC A
	case opcode == 0x0F:
		return rrc(state, regA), nil
	// RL r
	case opcode >= 0x10 && opcode <= 0x15:
		regIndex := opcode - 0x10
		return rl(state, indexToRegister[regIndex]), nil
	// RL (HL)
	case opcode == 0x16:
		return rlMemHL(state), nil
	// RL A
	case opcode == 0x17:
		return rl(state, regA), nil
	// RR A
	case opcode == 0x1F:
		return rr(state, regA), nil
	// RR (HL)
	case opcode == 0x1E:
		return rrMemHL(state), nil
	// RR r
	case opcode >= 0x18 && opcode <= 0x1D:
		regIndex := opcode - 0x18
		return rr(state, indexToRegister[regIndex]), nil
	// SLA r
	case opcode >= 0x20 && opcode <= 0x25:
		regIndex := opcode - 0x20
		return sla(state, indexToRegister[regIndex]), nil
	// SLA (HL)
	case opcode == 0x26:
		return slaMemHL(state), nil
	// SLA A
	case opcode == 0x27:
		return sla(state, regA), nil
	// SRA r
	case opcode >= 0x28 && opcode <= 0x2D:
		regIndex := opcode - 0x28
		return sra(state, indexToRegister[regIndex]), nil
	// SRA (HL)
	case opcode == 0x2E:
		return sraMemHL(state), nil
	// SRA A
	case opcode == 0x2F:
		return sra(state, regA), nil
	// SWAP r
	case opcode >= 0x30 && opcode <= 0x35:
		return swap(state, indexToRegister[lowerNibble]), nil
	// SWAP (HL)
	case opcode == 0x36:
		return swapMemHL(state), nil
	// SWAP A
	case opcode == 0x37:
		return swap(state, regA), nil
	// SRL a
	case opcode == 0x3F:
		return srl(state, regA), nil
	// SRL (HL)
	case opcode == 0x3E:
		return srlMemHL(state), nil
	// SRL r
	case opcode >= 0x38 && opcode <= 0x3D:
		regIndex := opcode - 0x38
		return srl(state, indexToRegister[regIndex]), nil
	// BIT b,A
	case upperNibble >= 0x04 && upperNibble <= 0x07 &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bitNum := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0F {
			bitNum++
		}

		return bit(state, bitNum, regA), nil
	// BIT b,(HL)
	case upperNibble >= 0x04 && upperNibble <= 0x07 &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x04) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return bitMemHL(state, bitNum), nil
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
		return bit(state, bitNum, indexToRegister[regIndex]), nil
	// RES b,A
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bit := (upperNibble - 0x08) * 2
		if lowerNibble == 0x0F {
			bit++
		}

		return res(state, bit, regA), nil
	// RES b,(HL)
	case upperNibble >= 0x08 && upperNibble <= 0x0B &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x08) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return resMemHL(state, bitNum), nil
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
		return res(state, bit, indexToRegister[regIndex]), nil
	// SET b,A
	case upperNibble >= 0x0C && upperNibble <= 0x0F &&
		(lowerNibble == 0x07 || lowerNibble == 0x0F):
		bit := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0F {
			bit++
		}

		return set(state, bit, regA), nil
	// SET b,(HL)
	case upperNibble >= 0x0C && upperNibble <= 0x0F &&
		(lowerNibble == 0x06 || lowerNibble == 0x0E):
		bitNum := (upperNibble - 0x0C) * 2
		if lowerNibble == 0x0E {
			bitNum++
		}

		return setMemHL(state, bitNum), nil
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
		return set(state, bit, indexToRegister[regIndex]), nil
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
