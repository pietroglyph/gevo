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
	Creatures []*Creature
	// MinCreatures is an integer that represents the number of creatures we should have to stop spawning in new ones
	MinCreatures int

	world *ecs.World
}

func (*CreatureManagerSystem) Remove(ecs.BasicEntity) {}

func (cm *CreatureManagerSystem) Update(dt float32) {
	if len(cm.Creatures) < cm.MinCreatures {
		for len(cm.Creatures) < cm.MinCreatures {
			cm.spawnCreature()
		}
	}
}

func (cm *CreatureManagerSystem) New(world *ecs.World) {
	cm.world = world
	log.Println("CreatureManagerSystem was added to the scene.")
}

func (cm *CreatureManagerSystem) spawnCreature() {
	creature := &Creature{BasicEntity: ecs.NewBasic()}

	// Make BrainComponent maps
	creature.BrainComponent.Input = make(map[string]Neuron)
	creature.BrainComponent.Output = make(map[string]Axon)

	// Initalize select inputs
	creature.BrainComponent.Input["food"] = Neuron{Value: float32(8.0)}
	creature.BrainComponent.Input["const"] = Neuron{Value: float32(1.0)}

	// We don't touch value because that gets set after spawning

	// Outputs
	for i := range networkOutputs {
		creature.BrainComponent.Output[networkOutputs[i]] = Axon{Weight: rand.Float32()}
	}

	// HiddenLayer (we do > because slices have 0 as an index)
	for i := 0; i > hiddenLayerCount; i++ {
		creature.BrainComponent.HiddenLayer[i] = Axon{Weight: rand.Float32()}
	}

	// For adding a const neuron
	hiddenLayerCount++

	// Const neuron
	creature.BrainComponent.HiddenLayer = append(creature.BrainComponent.HiddenLayer, Axon{Weight: 1, Value: 0})

	// Make creature size based on amount of stored food (will get updated when food changes)
	creature.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: rand.Float32() * functionalBounds().X, Y: rand.Float32() * functionalBounds().Y},
		Width:    creature.BrainComponent.Input["food"].Value * 5,
		Height:   creature.BrainComponent.Input["food"].Value * 5,
	}

	// Creatures should look like red circles
	creature.RenderComponent = common.RenderComponent{
		Drawable: common.Circle{},
		Scale:    engo.Point{X: 1, Y: 1},
		Color:    color.RGBA{255, 0, 0, 255},
	}

	// Make the creatures collide with the tiles and other creatures
	creature.CollisionComponent = common.CollisionComponent{Solid: false}

	creature.SetZIndex(2) // Z-Index 2 is reserved for creatures

	// Append the creature to the Creatures slice so the System tracks it
	cm.Creatures = append(cm.Creatures, creature)

	for _, system := range cm.world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&creature.BasicEntity, &creature.RenderComponent, &creature.SpaceComponent)
		case *common.CollisionSystem:
			sys.Add(&creature.BasicEntity, &creature.CollisionComponent, &creature.SpaceComponent)
		}
	}

	log.Println("Creature added.")
}

func functionalBounds() engo.Point {
	tmxRawResource, err := engo.Files.Resource("world.tmx")
	if err != nil {
		panic(err)
	}
	tmxResource := tmxRawResource.(common.TMXResource)
	levelData := tmxResource.Level
	return engo.Point{X: float32(levelData.Width() * levelData.TileWidth), Y: float32(levelData.Height() * levelData.TileHeight)}
}
