package main

import (
	"image/color"
	"log"
	"math"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/pietroglyph/gevo/chipecs"
	"github.com/pietroglyph/gevo/util"
	"github.com/vova616/chipmunk"
	"github.com/vova616/chipmunk/vect"
)

// MapScene satisfies the Scene interface
type MapScene struct {
	levelData    *common.Level
	tileEntities map[engo.Point]*tileEntity
}

// Label entity holds labels
type label struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// TileEntity holds all tilemap tiles
type tileEntity struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	chipecs.PhysicsComponent
	foodComponent
}

// FoodComponent holds all the tile's information relating to food
type foodComponent struct {
	waterDistance float32 // The distance in horizontal or vertical tiles from the current tile to a water tile (is 0 for water tiles)
	foodStored    float32 // Maxes out at (1 / waterDistance) * worldFertility, and goes lower when creature eats this tile
	deadly        bool    // Should creatures lose food when on this tile
}

var err error

var (
	scrollSpeed    float32 = 700.0
	zoomSpeed      float32 = -0.1
	worldFertility float32 = 1.5
)

// Type uniquely defines your game type
func (*MapScene) Type() string { return "map" }

// Preload is called before loading any assets from the disk,
// to allow you to register and queue them
func (*MapScene) Preload() {
	if err = engo.Files.Load("world.tmx"); err != nil { // Load tilemap
		panic(err)
	}
	if err = engo.Files.Load("AROLY.ttf"); err != nil { // Load logo font
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (ms *MapScene) Setup(world *ecs.World) {
	log.Println("Preloading map scene.")

	// Set the background color to green
	common.SetBackground(color.White)

	// Systems to make stuff actually happen in the world
	physicsSystem := &chipecs.PhysicsSystem{}
	world.AddSystem(&common.RenderSystem{})                                                                        // Render the game
	world.AddSystem(common.NewKeyboardScroller(scrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis)) // Use WASD to move the camera
	world.AddSystem(&common.MouseZoomer{ZoomSpeed: zoomSpeed})                                                     // Use the scrollwheel to zoom in and out
	world.AddSystem(physicsSystem)                                                                                 // Collide with stuff
	world.AddSystem(&CreatureManagerSystem{MapScene: ms, MinCreatures: 300})                                       // Add and manage creatures

	tmxRawResource, err := engo.Files.Resource("world.tmx")
	if err != nil {
		panic(err)
	}
	tmxResource := tmxRawResource.(common.TMXResource)
	ms.levelData = tmxResource.Level

	// Make the map for the holding the actual tile entities and extra data
	ms.tileEntities = make(map[engo.Point]*tileEntity, 0)

	// Set up camera Bounds
	common.CameraBounds = ms.levelData.Bounds()

	boundaries := []*chipmunk.Shape{
		chipmunk.NewSegment(util.PntToVect(ms.levelData.Bounds().Min), vect.Vect{X: vect.Float(ms.levelData.Bounds().Max.X), Y: vect.Float(0)}, vect.Float(0)),
		chipmunk.NewSegment(vect.Vect{X: vect.Float(ms.levelData.Bounds().Max.X), Y: vect.Float(0)}, util.PntToVect(ms.levelData.Bounds().Max), vect.Float(0)),
		chipmunk.NewSegment(util.PntToVect(ms.levelData.Bounds().Max), vect.Vect{X: vect.Float(0), Y: vect.Float(ms.levelData.Bounds().Max.Y)}, vect.Float(0)),
		chipmunk.NewSegment(vect.Vect{X: vect.Float(0), Y: vect.Float(ms.levelData.Bounds().Max.Y)}, util.PntToVect(ms.levelData.Bounds().Min), vect.Float(0)),
	}
	boundaryStaticBody := chipmunk.NewBodyStatic()
	for _, segment := range boundaries {
		segment.SetElasticity(0.6)
		segment.Shape().GetAsSegment().A.Sub(vect.Vect{X: vect.Float(ms.levelData.TileHeight), Y: vect.Float(ms.levelData.TileWidth)})
		segment.Shape().GetAsSegment().B.Sub(vect.Vect{X: vect.Float(ms.levelData.TileHeight), Y: vect.Float(ms.levelData.TileWidth)})
		boundaryStaticBody.AddShape(segment)
	}

	// Add all the actual tiles
	for _, tileLayer := range ms.levelData.TileLayers {
		for _, tileElement := range tileLayer.Tiles {
			if tileElement.Image != nil {
				tile := &tileEntity{BasicEntity: ecs.NewBasic()}

				switch tileLayer.Name {
				case "Water Layer":
					tile.RenderComponent.SetZIndex(1) // Functionally the same as Z-Index 0 because all creatures are Z-index 2
					tile.foodComponent.foodStored = 0 // We can't eat this
					tile.foodComponent.deadly = true  // Creatures will drown here
					tile.foodComponent.waterDistance = 0
				case "Food Layer":
					tile.RenderComponent.SetZIndex(0) // Lowest Z-Index but functionally the same as Z-Index 1
					// Loop over the the Water Layer and find the closest water tiles (not dependent on Water Layer entities existing)
					for _, layer := range ms.levelData.TileLayers {
						if layer.Name == "Water Layer" {
							var minDistance float32
							for _, t := range layer.Tiles {
								// We do all this to find an int representing the distance from a water tile to a food tile
								// We're basically normalizing a vector
								p := util.SubtractPoints(t.Point, tileElement.Point)
								dist := float32(math.Abs(float64(p.X/tileElement.Width())) + math.Abs(float64(p.Y/tileElement.Height())))
								// FIXME: Using t instead of tileElement causes a segfault, so we use tileElement instead... This could screw up if layers have different tile sizes
								if dist <= minDistance || minDistance == 0.0 { // Check if this is closer than any other tiles we've seen
									minDistance = dist
								}
								if minDistance == 1 { // The distance isn't going to be smaller than 1 so we can stop
									break
								}
							}
							// Actually set the values we've caluclated
							tile.foodComponent.waterDistance = minDistance
							tile.foodComponent.foodStored = (1 / minDistance) * worldFertility
						}
					}
					if tile.foodComponent.waterDistance == 0.0 { // This shouldn't happen unless the tilemap is screwed up
						log.Fatal("No Water Layer in tilemap")
					}
					tile.foodComponent.deadly = false // Food certainly isn't deadly
				}

				tile.RenderComponent = common.RenderComponent{
					Drawable: tileElement,
					Scale:    engo.Point{X: 1, Y: 1},
				}

				// Make the food tiles varying shades of green, based upon their foodStored
				if tileLayer.Name == "Food Layer" {
					mod := uint8((tile.foodComponent.foodStored / worldFertility) * 200)
					tile.RenderComponent.Color = color.RGBA{0, mod, 0, 255}
				}

				tile.SpaceComponent = common.SpaceComponent{
					Position: tileElement.Point,
					Width:    tileElement.Width(),
					Height:   tileElement.Height(),
				}

				_, exists := ms.tileEntities[tileElement.Point]
				if exists {
					log.Println("Overlapping tiles detected at", tileElement.Point)
				}
				ms.tileEntities[tileElement.Point] = tile
			}
		}
	}

	// Do the same for all image layers (there probably won't be any in this case)
	for _, imageLayer := range ms.levelData.ImageLayers {
		for _, imageElement := range imageLayer.Images {
			if imageElement.Image != nil {
				tile := &tileEntity{BasicEntity: ecs.NewBasic()}
				tile.RenderComponent = common.RenderComponent{
					Drawable: imageElement,
					Scale:    engo.Point{X: 1, Y: 1},
				}
				tile.SpaceComponent = common.SpaceComponent{
					Position: imageElement.Point,
					Width:    imageElement.Width(),
					Height:   imageElement.Height(),
				}

				ms.tileEntities[imageElement.Point] = tile
			}
		}
	}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			for _, v := range ms.tileEntities { // Add all of the tiles/imageLayers
				sys.Add(&v.BasicEntity, &v.RenderComponent, &v.SpaceComponent)
			}
		case *chipecs.PhysicsSystem:
			sys.Space.AddBody(boundaryStaticBody)
		}
	}
}

func (ms *MapScene) getTileEntityAt(p engo.Point) *tileEntity {
	closestTilePoint := engo.Point{}
	closestTilePoint.X = float32((int(p.X) / ms.levelData.TileWidth) * ms.levelData.TileWidth)
	closestTilePoint.Y = float32((int(p.Y) / ms.levelData.TileHeight) * ms.levelData.TileHeight)
	_, exists := ms.tileEntities[closestTilePoint]
	if !exists {
		log.Println("Get of a nonexistant tile at", closestTilePoint)
		return &tileEntity{foodComponent: foodComponent{deadly: true}} // Nonexistant tiles are deadly
	}
	return ms.tileEntities[closestTilePoint]
}
