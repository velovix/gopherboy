package gameboy

// call loads a 16-bit address, pushes the address of the next instruction onto
// the stack, and jumps to the loaded address.
func call(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read lower byte of destination address

		addressLower := state.incrementPC()

		return func(state *State) instruction {
			// M-Cycle 2: Read upper byte of destination address

			addressUpper := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 3: Some unknown internal operation

				return func(state *State) instruction {
					// M-Cycle 4: Push the upper byte of the current PC onto
					//            the stack

					_, pcUpper := split16(state.regPC.get())
					state.pushToStack(pcUpper)

					return func(state *State) instruction {
						// M-Cycle 5: Push the lower byte of the current PC
						//            onto the stack and set the PC to the
						//            target address

						pcLower, _ := split16(state.regPC.get())
						state.pushToStack(pcLower)

						// TODO(velovix): Is this actually where the PC is set?
						state.regPC.setLower(addressLower)
						state.regPC.setUpper(addressUpper)

						return nil
					}
				}
			}
		}

	}
}

// makeCALLIfFlag creates an instruction that loads a 16-bit address, pushes
// the address of the next instruction onto the stack, and jumps to the loaded
// address if the given flag is at the expected setting.
func makeCALLIfFlag(flagMask uint8, isSet bool) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Read lower byte of destination address

			addressLower := state.incrementPC()

			return func(state *State) instruction {
				// M-Cycle 2: Read upper byte of destination address and check
				//            the flag value

				addressUpper := state.incrementPC()

				flagState := state.regF.get()&flagMask == flagMask

				if flagState != isSet {
					// Condition evaluated to false, don't jump to the target
					// address
					return nil
				}

				return func(state *State) instruction {
					// M-Cycle 3: Some unknown internal operation

					return func(state *State) instruction {
						// M-Cycle 4: Push the upper byte of the current PC
						//            onto the stack

						_, pcUpper := split16(state.regPC.get())
						state.pushToStack(pcUpper)

						return func(state *State) instruction {
							// M-Cycle 5: Push the lower byte of the current PC
							//            onto the stack and set the PC to the
							//            target address

							pcLower, _ := split16(state.regPC.get())
							state.pushToStack(pcLower)

							// TODO(velovix): Is this actually where the PC is set?
							state.regPC.setLower(addressLower)
							state.regPC.setUpper(addressUpper)

							return nil
						}
					}
				}
			}
		}
	}
}

// ret pops a 16-bit address from the stack and jumps to it.
func ret(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Pop the lower byte of the target address off the stack

		addressLower := state.popFromStack()

		return func(state *State) instruction {
			// M-Cycle 2: Pop the upper byte of the target address off the
			//            stack

			addressUpper := state.popFromStack()

			return func(state *State) instruction {
				// M-Cycle 3: Probably set the PC to the target address

				state.regPC.setLower(addressLower)
				state.regPC.setUpper(addressUpper)

				return nil
			}
		}
	}
}

// makeRETIfFlag creates an instruction that pops a 16-bit address from the
// stack and jumps to it, but only if the given flag is at the expected value.
func makeRETIfFlag(flagMask uint8, isSet bool) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Check flag value

			flagState := state.regF.get()&flagMask == flagMask

			if flagState != isSet {
				// Condition evaluated to false, don't jump to the target
				// address
				return nil
			}

			return func(state *State) instruction {
				// M-Cycle 2: Read the lower byte of the target address

				addressLower := state.popFromStack()

				return func(state *State) instruction {
					// M-Cycle 3: Read the upper byte of the target address

					addressUpper := state.popFromStack()

					return func(state *State) instruction {
						// M-Cycle 4: Set the PC to the target address

						state.regPC.setLower(addressLower)
						state.regPC.setUpper(addressUpper)

						return nil
					}
				}
			}
		}
	}
}

// reti pops a 16-bit address from the stack and jumps to it, then enables
// interrupts.
func reti(state *State) instruction {
	// M-Cycle 0: Fetch instruction

	return func(state *State) instruction {
		// M-Cycle 1: Read lower byte of target address

		addressLower := state.popFromStack()

		return func(state *State) instruction {
			// M-Cycle 2: Read upper byte of target address

			addressUpper := state.popFromStack()

			return func(state *State) instruction {
				// M-Cycle 3: Set the PC to the target address and turn on
				//            interrupts

				state.regPC.setLower(addressLower)
				state.regPC.setUpper(addressUpper)

				state.interruptsEnabled = true

				return nil
			}
		}
	}
}

// makeRST creates an instruction that pushes the current program counter to
// the stack and jumps to the given address.
func makeRST(address uint16) instruction {
	return func(state *State) instruction {
		// M-Cycle 0: Fetch instruction

		return func(state *State) instruction {
			// M-Cycle 1: Some unknown internal behavior

			return func(state *State) instruction {
				// M-Cycle 2: Push the upper byte of the PC to the stack

				_, upper := split16(state.regPC.get())
				state.pushToStack(upper)

				return func(state *State) instruction {
					// M-Cycle 3: Push the lower byte of the PC to the stack
					//            and set PC to target address

					lower, _ := split16(state.regPC.get())
					state.pushToStack(lower)

					// TODO(velovix): Is this where the PC is actually set?
					state.regPC.set(address)

					return nil
				}
			}
		}
	}
}
