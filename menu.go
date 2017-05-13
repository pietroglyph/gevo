package main

import (
	"image/color"

	"engo.io/ecs"
	"engo.io/engo"
	"engo.io/engo/common"
)

type menuScene struct{}

type label struct {
	ecs.BasicEntity
	common.RenderComponent
	common.SpaceComponent
}

// Type uniquely defines your game type
func (*menuScene) Type() string { return "menu" }

// Preload is called before loading any assets from the disk,
// to allow you to register / queue them
func (*menuScene) Preload() {
	err := engo.Files.Load("AROLY.ttf")
	if err != nil {
		panic(err)
	}
}

// Setup is called before the main loop starts. It allows you
// to add entities and systems to your Scene.
func (*menuScene) Setup(world *ecs.World) {
	common.SetBackground(color.RGBA{0, 204, 0, 1})
	world.AddSystem(&common.RenderSystem{})

	fnt := &common.Font{
		URL:  "AROLY.ttf",
		FG:   color.White,
		Size: 128,
	}

	err := fnt.CreatePreloaded()
	if err != nil {
		panic(err)
	}

	titleLabel := label{BasicEntity: ecs.NewBasic()}
	titleLabel.RenderComponent.Drawable = common.Text{
		Font: fnt,
		Text: "gevo",
	}
	titleLabel.SpaceComponent.Center()

	for _, system := range world.Systems() {
		switch sys := system.(type) {
		case *common.RenderSystem:
			sys.Add(&titleLabel.BasicEntity, &titleLabel.RenderComponent, &titleLabel.SpaceComponent)
		}
	}
}
