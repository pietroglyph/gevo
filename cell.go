package main

import tl "github.com/JoelOtter/termloop"

// Cell is a single entity, with instructions, and stored energy (food); a cell can reproduce, move, and eat chronologically based on these instructions and its internal state
type Cell struct {
	*tl.Entity
	Food int           // Food is used for every action that the cell preforms, and is gained by eating (other cells, or the slime)
	Inst []Instruction // Genes are stored as a slice of a Gene, they can be of arbitrary length
	Name rune          // Name is a single unique code point assigned to this cell and all of its *identical* offspring
}

// TODO: Add entity drawing, collision, and interaction code
