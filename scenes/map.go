package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type mapScene struct{}

type label struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

type tile struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
	food int
}

var err error

// Type uniquely defines your game type
func (*mapScene) Type() string { return "map" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*mapScene) Preload() {
	err = engo.Files.Load("AROLY.ttf") // Load the font for the game title label
	if err != nil {
		panic(err)
	}
	if err := engo.Files.Load("world.tmx"); err != nil { // Load a tilemap for creatures to live on
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (*mapScene) Setup(world *ecs.World) {
	common.SetBackground(color.RGBA{0, 168, 0, 1})
	world.AddSystem(&common.RenderSystem{})

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

				tile := &tile{BasicEntity: ecs.NewBasic()}
				tile.RenderComponent = common.RenderComponent{
					Drawable: tileElement,
					Scale:    engo.Point{1, 1},
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
					Scale:    engo.Point{1, 1},
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
