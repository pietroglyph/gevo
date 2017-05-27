package main

import (
	"github.com/pietroglyph/gevo/scenes"

	"engo.io/engo"
)

func main() {
	opts := engo.RunOptions{
		Title:          "gevo",
		Width:          1600,
		Height:         1600,
		StandardInputs: true,
		MSAA:           0,
	}
	engo.Run(opts, &scenes.MapScene{})
}
