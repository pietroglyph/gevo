package scenes

import (
	"fmt"
	"image/color"
	"log"
	"math"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
	"github.com/pietroglyph/gevo/systems"
	"github.com/pietroglyph/gevo/util"
)

// MapScene satisfies the Scene interface
type MapScene struct{}

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
	common.CollisionComponent
	foodComponent
}

// TileComponent holds all the tile's information relating to food
type foodComponent struct {
	waterDistance float64 // The distance in horizontal or vertical tiles from the current tile to a water tile (is 0 for water tiles)
	foodStored    float64 // Maxes out at (1/waterDistance) * worldFertility, and goes lower when creature eats this tile
	deadly        bool    // Should creatures lose food when on this tile
}

var err error

var (
	scrollSpeed    float32 = 700.0
	zoomSpeed      float32 = -0.1
	worldFertility float64 = 4.0
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
func (*MapScene) Setup(world *ecs.World) {
	log.Println("Preloading map scene.")
	// Set the background color to green
	common.SetBackground(color.White)

	// Systems to make stuff actually happen in the world
	world.AddSystem(&common.RenderSystem{})                                                                        // Render the game
	world.AddSystem(common.NewKeyboardScroller(scrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis)) // Use WASD to move the camera
	world.AddSystem(&common.MouseZoomer{zoomSpeed})                                                                // Use the scrollwheel to zoom in and out
	world.AddSystem(&common.CollisionSystem{})                                                                     // Collide with stuff
	world.AddSystem(&systems.CreatureManagerSystem{MinCreatures: 10000})                                           // Add and manage creatures
	arolyFont := &common.Font{
		URL:  "AROLY.ttf",
		FG:   color.White,
		Size: 128,
	}

	err = arolyFont.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	titleLabel := label{BasicEntity: ecs.NewBasic()}
	titleLabel.RenderComponent.Drawable = common.Text{
		Font: arolyFont,
		Text: "gevo",
	}

	tmxRawResource, err := engo.Files.Resource("world.tmx")
	if err != nil {
		panic(err)
	}
	tmxResource := tmxRawResource.(common.TMXResource)
	levelData := tmxResource.Level

	// Create render and space components for each of the tiles in all layers
	tileComponents := make([]*tileEntity, 0)

	// Add all the actual tiles
	for _, tileLayer := range levelData.TileLayers {
		for _, tileElement := range tileLayer.Tiles {
			if tileElement.Image != nil {
				tile := &tileEntity{BasicEntity: ecs.NewBasic()}

				switch tileLayer.Name {
				case "Collision Layer":
					tile.RenderComponent.SetZIndex(3)    // Highest Z-Index, everything will pass under it
					tile.foodComponent.foodStored = 0    // We can't eat this
					tile.foodComponent.waterDistance = 0 // This value doesn't really matter for this layer
					tile.foodComponent.deadly = false    // Creatures should collide with this tile, so this doesn't really matter
					tile.CollisionComponent = common.CollisionComponent{Solid: true}
				case "Water Layer":
					tile.RenderComponent.SetZIndex(1) // Functionally the same as Z-Index 0 because all creatures are Z-index 2
					tile.foodComponent.foodStored = 0 // We can't eat this
					tile.foodComponent.deadly = true  // Creatures will drown here
					tile.foodComponent.waterDistance = 0
					tile.CollisionComponent = common.CollisionComponent{Solid: false}
				case "Food Layer":
					tile.RenderComponent.SetZIndex(0) // Lowest Z-Index but functionally the same as Z-Index 1
					// Loop over the the Water Layer and find the closest water tiles (not dependent on Water Layer entities existing)
					for _, layer := range levelData.TileLayers {
						if layer.Name == "Water Layer" {
							var minDistance float64
							for _, t := range layer.Tiles {
								// We do all this to find an int representing the distance from a water tile to a food tile
								// We're basically normalizing a vector
								p := util.SubtractPoints(t.Point, tileElement.Point)                                             // Point
								dist := math.Abs(float64(p.X/tileElement.Width())) + math.Abs(float64(p.Y/tileElement.Height())) // Float64
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
						panic(fmt.Errorf("No Water Layer in tilemap!"))
					}
					tile.foodComponent.deadly = false // Food certainly isn't deadly
					tile.CollisionComponent = common.CollisionComponent{Solid: false}
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

				tileComponents = append(tileComponents, tile)
			}
		}
	}

	// Do the same for all image layers (there probably won't be any in this case)
	for _, imageLayer := range levelData.ImageLayers {
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

				tileComponents = append(tileComponents, tile)
			}
		}
	}

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&titleLabel.BasicEntity, &titleLabel.RenderComponent, &titleLabel.SpaceComponent) // Add the game title label
			for _, v := range tileComponents {                                                        // Add all of the tiles/imageLayers
				sys.Add(&v.BasicEntity, &v.RenderComponent, &v.SpaceComponent)
			}
		case *common.CollisionSystem:
			for _, v := range tileComponents {
				sys.Add(&v.BasicEntity, &v.CollisionComponent, &v.SpaceComponent)
			}
		}
	}
}
