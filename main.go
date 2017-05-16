package main

import "engo.io/engo"

func main() {
	opts := engo.RunOptions{
		Title:  "gevo",
		Width:  1920,
		Height: 1080,
	}
	engo.Run(opts, &mapScene{})
}
