package main

import (
	"engo.io/engo"
)

func main() {
	opts := engo.RunOptions{
		Title:          "gevo",
		Width:          800,
		Height:         800,
		StandardInputs: true,
		MSAA:           3,
		VSync:          true,
		Fullscreen:     false,
		ScaleOnResize:  false,
		NotResizable:   true,
	}
	engo.Run(opts, &MapScene{})
}
