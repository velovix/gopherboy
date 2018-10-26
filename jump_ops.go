package main

// jr loads a signed offset value, then jumps to the operation at address PC +
// offset. In other words, it's a jump relative to the current position.
func jr(env *environment) int {
	offset := asSigned(env.incrementPC())
	env.relativeJump(int(offset))

	return 12
}

// jrIfFlag loads an offset value, then jumps to the operation at address PC +
// offset if the given flag is at the expected setting.
func jrIfFlag(env *environment, flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask
	offset := asSigned(env.incrementPC())

	//conditional := getConditionalStr(flagMask, isSet)
	//fmt.Printf("JR %v,%#x\n", conditional, offset)
	if flagState == isSet {
		env.relativeJump(int(offset))
		return 12
	} else {
		return 8
	}
}

// jp loads a 16-bit address and jumps to it.
func jp(env *environment) int {
	address := combine(env.incrementPC(), env.incrementPC())
	env.regs[regPC].set(address)

	return 16
}

// jpIfFlag loads a 16-bit address and jumps to it if the given flag is at the
// expected setting.
func jpIfFlag(env *environment, flagMask uint16, isSet bool) int {
	flagState := env.regs[regF].get()&flagMask == flagMask
	address := combine(env.incrementPC(), env.incrementPC())

	//conditional := getConditionalStr(flagMask, isSet)
	//fmt.Printf("JP %v,%#x\n", conditional, offset)
	if flagState == isSet {
		env.regs[regPC].set(address)
		return 16
	} else {
		return 12
	}
}
