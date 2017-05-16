package main

import (
	"engo.io/ecs"
	"engo.io/engo/common"
)

type creature struct {
	ecs.BasicEntity
	common.SpaceComponent
	common.RenderComponent
	brainComponent
}

type neuron struct {
	value float32
}

type axon struct {
	value  float32
	weight float32
}

type brainComponent struct {
	input map[string]neuron
}
