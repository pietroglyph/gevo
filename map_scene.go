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
	err = engo.Files.Load("AROLY.ttf")
	if err != nil {
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

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&titleLabel.BasicEntity, &titleLabel.RenderComponent, &titleLabel.SpaceComponent)
		}
	}
}
