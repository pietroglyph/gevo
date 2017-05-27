package scenes

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
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
	data tileData
}

// TileData provides an interface that gives all the information we need about tile
type tileData interface {
	getFood() int
	isSolid() bool
	getColorTransform() color.RGBA
}

var err error

var (
	scrollSpeed float32 = 700
	zoomSpeed   float32 = -0.1
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
	// Set the background color to green
	common.SetBackground(color.Black)

	// Systems to make stuff actually happen in the world
	world.AddSystem(&common.RenderSystem{})                                                                        // Render the game
	world.AddSystem(common.NewKeyboardScroller(scrollSpeed, engo.DefaultHorizontalAxis, engo.DefaultVerticalAxis)) // Use WASD to move the camera
	world.AddSystem(&common.MouseZoomer{zoomSpeed})                                                                // Use the scrollwheel to zoom in and out

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
	tileComponents := make([]*tile, 0)

	for _, tileLayer := range levelData.TileLayers {
		for _, tileElement := range tileLayer.Tiles {
			if tileElement.Image != nil {
				tile := &tile{BasicEntity: ecs.NewBasic(), food: 4} // FIXME: Find a way to figure out the type of tile, and give it an actual food value
				tile.RenderComponent = common.RenderComponent{
					Drawable: tileElement,
					Scale:    engo.Point{X: 1, Y: 1},
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
				tile := &tile{BasicEntity: ecs.NewBasic()}
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
		}
	}
}
