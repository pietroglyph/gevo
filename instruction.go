package main

// Instruction is an atomic action, with a type and a parameter
type Instruction struct {
	Type      int // Type is a type of action represented as an integer
	Parameter int // Each action takes an integer as a parameter
}

// TODO: Improve implementation, possibly have a type for each instruction and have them all satisfy an "instruction" interface
