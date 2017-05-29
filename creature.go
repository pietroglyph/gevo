package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

// Creature is an entity upon which evolution is simulated
// Creatures can collide, have a size, have a sprite,
// and also have a "brain" which is a very simple 2-layer feedforward neural network.
// The weights of this network are analogous to genetic information.
type Creature struct {
	ecs.BasicEntity
	common.SpaceComponent
	common.RenderComponent
	common.CollisionComponent
	// BrainComponent contains a simple feedforward neural network
	BrainComponent
}

// Neuron has a single value field, and is meant to be used as an input
// Thus, it is unweighted
type Neuron struct {
	// Value is the unweighted values of the neuron
	Value float32
}

// Axon has a value, and a weight, it is intended to be used in all but the input layers
type Axon struct {
	// Value is the value, with the weight applied to it
	Value float32
	// Weight is the value we should apply to Value
	Weight float32
}

// Brain component contains a simple 2-layer feedforward neural network
type BrainComponent struct {
	// Input is a map of unweighted values
	Input map[string]Neuron
	// HiddenLayer is a map of weighted values, the key corresponds to an Input key
	HiddenLayer []Axon
	// Output is a map of weighted values, the key corresponds to a HiddenLayer key
	Output map[string]Axon
}
