package main

import (
	"image/color"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/pietroglyph/gevo/chipecs"
	"github.com/pietroglyph/gevo/util"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

var (
	networkInputs                  = []string{"rotation", "storedfood", "vision", "const"}
	networkOutputs                 = []string{"velocitydelta", "angledelta", "eat", "mate"}
	hiddenLayerCount               = len(networkInputs) + len(networkOutputs)
	creatureSizeMultiplier float32 = 4.0
	massMultiplier         float32 = 5
	baseFoodCost           float32 = 0.3
	movementFoodCost       float32 = 0.4
	rotationFoodCost       float32 = 0.1
	eatFoodCost            float32 = 0.2
	deadlyTileFoodCost     float32 = 10
	wg                     sync.WaitGroup
	elapsedTime            int
)

// Creature is an entity upon which evolution is simulated
// Creatures can collide, have a size, and something to render,
// and also have a "brain" which is a very simple 2-layer feedforward neural network.
// The weights of this network are analogous to genetic information.
type Creature struct {
	ecs.BasicEntity
	common.SpaceComponent
	common.RenderComponent
	chipecs.PhysicsComponent
	// BrainComponent contains a simple feedforward neural network
	BrainComponent
	StoredFood float32
}

// Neuron has a single value field, and is meant to be used as an input
// Thus, it is unweighted
type Neuron struct {
	// Value is the unweighted value of the neuron
	Value float32
}

// Axon has a value, and a weight, it is intended to be used in all but the input layers
type Axon struct {
	// Value is the value, with the weight applied to it
	Value float32
	// Weight is the value we should apply to Value
	Weight float32
}

// BrainComponent contains a simple 2-layer feedforward neural network
type BrainComponent struct {
	// Input is a map of unweighted values
	Input map[string]Neuron
	// HiddenLayer is a map of weighted values, the key corresponds to an Input key
	HiddenLayer []Axon
	// Output is a map of weighted values, the key corresponds to a HiddenLayer key
	Output map[string]Axon
}

// CreatureManagerSystem satisfies interface ecs.System
type CreatureManagerSystem struct {
	// Creatures is a Creature slice containing all the creatures in the World that should be managed
	Creatures map[uint64]*Creature
	// MinCreatures is an integer that represents the number of creatures we should have to stop spawning in new ones
	MinCreatures int
	// MapScene holds a pointer to the map scene
	MapScene *MapScene

	// World is used to keep track of game's world because we need it in update
	World *ecs.World
}

func (c *Creature) think(ms *MapScene) {
	defer wg.Done() // Decrement the WaitGroup when we're done

	// Populate Input
	for key := range c.BrainComponent.Input {
		// We do this because doing c.BrainComponent.Input[key].Value is a double assignment if key doesn't exits, which Go doesn't allow
		var val = c.BrainComponent.Input[key] // We're making a copy here where we first assume that key exists
		switch key {
		case "rotation":
			val.Value = c.Rotation
		case "storedfood":
			val.Value = c.StoredFood
		case "vision":
			val.Value = ms.getTileEntityAt(c.Position).foodStored
		case "const":
			val.Value = 1
		}
		c.BrainComponent.Input[key] = val
	}

	// Populate HiddenLayer
	for i := range c.BrainComponent.HiddenLayer {
		var wSum float32
		// Find the weighted sum of the Input layer
		for key := range c.BrainComponent.Input {
			wSum += c.BrainComponent.Input[key].Value * c.BrainComponent.HiddenLayer[i].Weight
		}
		c.BrainComponent.HiddenLayer[i].Value = wSum
	}

	// Populate Output
	for key := range c.BrainComponent.Output {
		var wSum float32
		// Find the weighted sum of the HiddenLayer
		for i := range c.BrainComponent.HiddenLayer {
			wSum += c.BrainComponent.HiddenLayer[i].Value * c.BrainComponent.Output[key].Weight
		}
		// See the first loop for why we do this
		var val = c.BrainComponent.Output[key]
		val.Value = wSum
		c.BrainComponent.Output[key] = val
	}
	return
}

// Remove is called when an entity is removed
func (cm *CreatureManagerSystem) Remove(e ecs.BasicEntity) {
	delete(cm.Creatures, e.ID())
}

// Update is called every frame
func (cm *CreatureManagerSystem) Update(dt float32) {
	if len(cm.Creatures) < cm.MinCreatures {
		for len(cm.Creatures) < cm.MinCreatures {
			cm.spawnCreature()
		}
	}

	for _, v := range cm.Creatures {
		wg.Add(1)
		go v.think(cm.MapScene)
	}
	wg.Wait()

	for _, v := range cm.Creatures {
		// Update the current position and rotation based on the angle and position delta
		v.Body.AddAngle(v.Output["angledelta"].Value)
		v.Body.AddAngularVelocity(v.Output["velocitydelta"].Value)
		// Use food for everything that's being done, and eat
		v.StoredFood -= v.Output["angledelta"].Value * rotationFoodCost
		v.StoredFood -= v.Output["movementdelta"].Value * movementFoodCost
		v.StoredFood -= baseFoodCost
		if v.Output["eat"].Value > 0 {
			v.StoredFood -= v.Output["eat"].Value * eatFoodCost
			tileUnder := cm.MapScene.getTileEntityAt(v.SpaceComponent.Center())
			v.StoredFood += float32(tileUnder.foodStored)
			if tileUnder.deadly {
				v.StoredFood -= deadlyTileFoodCost
			}
		}
		if v.StoredFood < 0.3 {
			cm.World.RemoveEntity(v.BasicEntity)
		}
		diameter := v.StoredFood * creatureSizeMultiplier
		v.Width = diameter
		v.Height = diameter
	}
}

// New is called when CreatureManagerSystem is added to the scene
func (cm *CreatureManagerSystem) New(World *ecs.World) {
	cm.World = World                          // So we can access World in cm.Update
	rand.Seed(time.Now().UnixNano())          // Use the current Unix time as a seed for our random numbers
	cm.Creatures = make(map[uint64]*Creature) // Make the Creatures map

	engo.Mailbox.Listen("CollisionMessage", func(message engo.Message) {
		m, ok := message.(common.CollisionMessage)
		if !ok {
			return
		}

		_, fromExists := cm.Creatures[m.Entity.ID()]
		_, toExists := cm.Creatures[m.To.ID()]
		if !fromExists || !toExists {
			return
		}
		if cm.Creatures[m.Entity.ID()].Output["mate"].Value > 5 && cm.Creatures[m.To.ID()].Output["mate"].Value > 5 {
			if rand.Float64() < 0.99 {
				return
			}
			cm.spawnCreature() // TODO: Add genetic inheritance
		} else {
			if cm.Creatures[m.Entity.ID()].StoredFood > cm.Creatures[m.To.ID()].StoredFood {
				cm.Creatures[m.To.ID()].StoredFood -= cm.Creatures[m.To.ID()].StoredFood
			}
		}
	})
	log.Println("CreatureManagerSystem was added to the scene.")
}

func (cm *CreatureManagerSystem) spawnCreature() {
	rand.Seed(time.Now().UnixNano())
	creature := &Creature{BasicEntity: ecs.NewBasic()}

	// Make BrainComponent maps
	creature.BrainComponent.Input = make(map[string]Neuron)
	creature.BrainComponent.Output = make(map[string]Axon)

	// Initalize select inputs
	creature.StoredFood = 8
	creature.BrainComponent.Input["food"] = Neuron{Value: creature.StoredFood}
	creature.BrainComponent.Input["const"] = Neuron{Value: float32(1.0)}

	// We don't touch Value because that gets set after spawning

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

	bounds := engo.Point{X: float32(cm.MapScene.levelData.Width() * cm.MapScene.levelData.TileWidth), Y: float32(cm.MapScene.levelData.Height() * cm.MapScene.levelData.TileHeight)}

	// For calculating size based on food
	diameter := creature.StoredFood * creatureSizeMultiplier

	// Make creature size based on amount of stored food and put the creature at 0, 0 (we'll get a random position later)
	creature.SpaceComponent = common.SpaceComponent{
		Position: engo.Point{X: rand.Float32(), Y: rand.Float32()},
		Width:    diameter,
		Height:   diameter,
	}

	// This stops overlap but pushes creatures to the center... FIXME?
	if creature.SpaceComponent.Position.X < 0.5 { // If we're closer to the left and top walls then make sure the creatures aren't colliding with the walls
		creature.SpaceComponent.Position.X *= bounds.X                                 // Regular World bounds
		creature.SpaceComponent.Position.X += float32(cm.MapScene.levelData.TileWidth) // Make sure we don't intersect with the top or left walls
	} else { // Same but for the bottom and right walls (and the middle)
		creature.SpaceComponent.Position.X *= bounds.X - float32(cm.MapScene.levelData.TileWidth) - diameter // Make sure we can't intersect with the bottom or right walls
	}

	if creature.SpaceComponent.Position.Y < 0.5 { // If we're closer to the left and top walls then make sure the creatures aren't colliding with the walls
		creature.SpaceComponent.Position.Y *= bounds.Y                                  // Regular World bounds
		creature.SpaceComponent.Position.Y += float32(cm.MapScene.levelData.TileHeight) // Make sure we don't intersect with the top or left walls
	} else { // Same but for the bottom and right walls (and the middle)
		creature.SpaceComponent.Position.Y *= bounds.Y - float32(cm.MapScene.levelData.TileHeight) - diameter // Make sure we can't intersect with the bottom or right walls
	}

	// Creatures should look like red circles
	creature.RenderComponent = common.RenderComponent{
		Drawable: common.Circle{},
		Scale:    engo.Point{X: 1, Y: 1},
		Color:    color.RGBA{255, 0, 0, 255},
	}

	// Setup physics
	shape := chipmunk.NewCircle(vect.Vector_Zero, diameter/2)
	shape.SetElasticity(0.95)
	shape.SetFriction(50)

	mass := calculateMass(diameter)
	body := chipmunk.NewBody(mass, shape.Moment(float32(mass)))
	body.SetPosition(util.PntToVect(creature.Position))
	body.SetAngle(vect.Float(creature.SpaceComponent.Rotation))
	body.AddShape(shape)

	creature.PhysicsComponent = chipecs.PhysicsComponent{Body: body}

	creature.SetZIndex(2) // Z-Index 2 is reserved for creatures

	// Append the creature to the Creatures slice so the System tracks it
	cm.Creatures[creature.ID()] = creature

	for _, system := range cm.World.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&creature.BasicEntity, &creature.RenderComponent, &creature.SpaceComponent)
		case *chipecs.PhysicsSystem:
			sys.Add(&creature.BasicEntity, &creature.PhysicsComponent, &creature.SpaceComponent)
		}
	}
}

func calculateMass(diameter float32) vect.Float {
	return vect.Float(diameter * massMultiplier)
}
