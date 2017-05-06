package main

import (
	"log"

	tl "github.com/JoelOtter/termloop"
)

func main() {
	log.Println("Starting gevo...")
	game := tl.NewGame()
	level := tl.NewBaseLevel(tl.Cell{
		Bg: tl.ColorBlue,
		Fg: tl.ColorWhite,
		Ch: '☐',
	})
	game.Screen().SetLevel(level)
	game.Start()
}
