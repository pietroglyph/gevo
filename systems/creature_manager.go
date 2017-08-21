package systems

import (
	"image/color"
	"log"
	"math"
	"math/rand"
	"sync"
	"time"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

var (
	networkInputs                  = []string{"rotation", "storedfood", "vision", "const"}
	networkOutputs                 = []string{"positiondelta", "rotationdelta", "eat", "mate"}
	hiddenLayerCount               = len(networkInputs) + len(networkOutputs)
	creatureSizeMultiplier float32 = 5.0
	movementFoodCost       float32 = 0.002
	rotationFoodCost       float32 = 0.0005
	eatFoodCost            float32 = 0.0003
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
	common.CollisionComponent
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
	// Creatures is a Creature slice containing all the creatures in the world that should be managed
	Creatures []*Creature
	// MinCreatures is an integer that represents the number of creatures we should have to stop spawning in new ones
	MinCreatures int
	// PositionLine is used for resolving creature vision and rotation
	PositionLine *engo.Line

	creaturesMux sync.Mutex
	world        *ecs.World
}

func (c *Creature) think(cmSys *CreatureManagerSystem) {
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
			val.Value = 0 // FIXME
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

func (cm *CreatureManagerSystem) Remove(e ecs.BasicEntity) {
	cm.creaturesMux.Lock()
	defer cm.creaturesMux.Unlock()
	k := 0
	for i, n := range cm.Creatures {
		if cm.Creatures[i].ID() != e.ID() {
			cm.Creatures[k] = n
			k++
		}
	}
	cm.Creatures = cm.Creatures[:k]
}

func (cm *CreatureManagerSystem) Update(dt float32) {
	if len(cm.Creatures) < cm.MinCreatures {
		for len(cm.Creatures) < cm.MinCreatures {
			cm.spawnCreature()
		}
	}
	for _, v := range cm.Creatures {
		wg.Add(1)
		go v.think(cm)
	}
	wg.Wait()
	for _, v := range cm.Creatures {
		// Update the current position and rotation based on the angle and position delta
		delta := engo.Point{}
		v.SpaceComponent.Rotation = addDegrees(v.SpaceComponent.Rotation, v.Output["rotationdelta"].Value)
		delta.X = float32(math.Sin(float64(v.SpaceComponent.Rotation))) * (v.Output["positiondelta"].Value)
		delta.Y = float32(math.Cos(float64(v.SpaceComponent.Rotation))) * (v.Output["positiondelta"].Value)
		v.SpaceComponent.Position.Add(delta)
		// Use food for everything that's being done, and eat
		v.StoredFood -= v.Output["rotationdelta"].Value * rotationFoodCost
		v.StoredFood -= v.Output["movementdelta"].Value * movementFoodCost
		if v.Output["eat"].Value > 0 {
			v.StoredFood -= v.Output["eat"].Value * eatFoodCost
			v.StoredFood = 10
		}
		if v.StoredFood < 0.3 {
			cm.Remove(v.BasicEntity)
		}
		diameter := v.StoredFood * creatureSizeMultiplier
		v.Width = diameter
		v.Height = diameter
	}
}

func (cm *CreatureManagerSystem) New(world *ecs.World) {
	cm.world = world
	rand.Seed(time.Now().UTC().UnixNano()) // Use the current Unix time as a seed for our random numbers
	engo.Mailbox.Listen("CollisionMessage", func(message engo.Message) {
		t, isCollision := message.(common.CollisionMessage)
		log.Println("test")
		if isCollision {
			log.Println("DEAD", t.Entity)
		}
	})
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

	// For discovering world bounds
	tmxRawResource, err := engo.Files.Resource("world.tmx")
	if err != nil {
		log.Panic(err.Error())
	}
	tmxResource := tmxRawResource.(common.TMXResource)
	levelData := tmxResource.Level
	bounds := engo.Point{X: float32(levelData.Width() * levelData.TileWidth), Y: float32(levelData.Height() * levelData.TileHeight)}

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
		creature.SpaceComponent.Position.X *= bounds.X                     // Regular world bounds
		creature.SpaceComponent.Position.X += float32(levelData.TileWidth) // Make sure we don't intersect with the top or left walls
	} else { // Same but for the bottom and right walls (and the middle)
		creature.SpaceComponent.Position.X *= bounds.X - float32(levelData.TileWidth) - diameter // Make sure we can't intersect with the bottom or right walls
	}

	if creature.SpaceComponent.Position.Y < 0.5 { // If we're closer to the left and top walls then make sure the creatures aren't colliding with the walls
		creature.SpaceComponent.Position.Y *= bounds.Y                      // Regular world bounds
		creature.SpaceComponent.Position.Y += float32(levelData.TileHeight) // Make sure we don't intersect with the top or left walls
	} else { // Same but for the bottom and right walls (and the middle)
		creature.SpaceComponent.Position.Y *= bounds.Y - float32(levelData.TileHeight) - diameter // Make sure we can't intersect with the bottom or right walls
	}

	// Creatures should look like red circles
	creature.RenderComponent = common.RenderComponent{
		Drawable: common.Circle{},
		Scale:    engo.Point{X: 1, Y: 1},
		Color:    color.RGBA{255, 0, 0, 255},
	}

	// Make the creatures collide with the tiles and other creatures
	creature.CollisionComponent = common.CollisionComponent{Solid: true}

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
}

func addDegrees(degrees float32, delta float32) float32 {
	degrees += delta
	if degrees > 360 {
		factorOver := int(math.Ceil(float64(degrees))) / 360
		degrees = degrees - float32(factorOver*360)
	} else if degrees < 0 {
		factorUnder := int(math.Ceil(math.Abs(float64(degrees)))) / 360
		degrees += float32(factorUnder+1) * 360
	}
	return degrees
}
