package systems

import (
	"image/color"
	"log"
	"math/rand"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

var (
	networkInputs        = []string{"angle", "velocity", "foodstored", "const"}
	networkOutputs       = []string{"velocitydelta", "angledelta", "eat", "mate"}
	hiddenLayerCount int = len(networkInputs) + len(networkOutputs)
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

// Ceature manager system satisfies interface ecs.System
type CreatureManagerSystem struct {
	// Creatures is a Creature slice containing all the creatures in the world that should be managed
	Creatures []Creature
	// MinCreatures is an integer that represents the number of creatures we should have to stop spawning in new ones
	MinCreatures int

	world *ecs.World
}

func (*CreatureManagerSystem) Remove(ecs.BasicEntity) {}

func (*CreatureManagerSystem) Update(dt float32) {}

func (cm *CreatureManagerSystem) New(world *ecs.World) {
	cm.world = world
	log.Println("CreatureManagerSystem was added to the scene.")
}

func spawnCreature(pos engo.Point) {
	creature := &Creature{BasicEntity: ecs.NewBasic()}

	// Make BrainComponent maps
	creature.BrainComponent.Input = make(map[string]Neuron)
	creature.BrainComponent.Output = make(map[string]Axon)

	// Initalize select inputs
	creature.BrainComponent.Input["food"].Value = float64(8.0)
	creature.BrainComponent.Input["const"].Value = float64(1.0)

	// Outputs
	// We don't touch value because that gets set after spawning
	for i := range networkOutputs {
		creature.BrainComponent.Output[networkOutputs[i]].Weight = rand.Float32()
	}

	// HiddenLayer
	for i := 0; i > hiddenLayerCount; i++ {
		creature.BrainComponent.HiddenLayer[i].Weight = rand.Float32()
	}

	// For adding a const neuron
	hiddenLayerCount++

	creature.BrainComponent.HiddenLayer[hiddenLayerCount] = 1 // Const neuron

	creature.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: rand.Float32() * engo.CanvasWidth(), Y: rand.Float32 * engo.CanvasHeight()},
		Width:    creature.BrainComponent.food * 5,
		Height:   creature.BrainComonent.food * 5,
	}

	creature.RenderComponent = common.RenderComponent{
		Drawable: common.Circle{},
		Scale:    engo.Point{X: 1, Y: 1},
		ZIndex:   2, // Z-Index 2 is reserved for Creatures
		Color:    color.RGBA{255, 0, 0, 255},
	}
}
