package main

import (
	"engo.io/engo"
)

func main() {
	opts := engo.RunOptions{
		Title:          "gevo",
		Width:          1600,
		Height:         1600,
		StandardInputs: true,
		MSAA:           3,
		VSync:          true,
		Fullscreen:     false,
	}
	engo.Run(opts, &MapScene{})
}
