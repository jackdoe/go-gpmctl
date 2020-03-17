package main

import (
	"log"

	gpmctl "github.com/jackdoe/go-gpmctl"
)

func main() {
	g, err := gpmctl.NewGPM(gpmctl.GPMConnect{
		EventMask:   gpmctl.ANY,
		DefaultMask: ^gpmctl.HARD,
		MinMod:      0,
		MaxMod:      ^uint16(0),
	})

	if err != nil {
		panic(err)
	}
	for {
		event, err := g.Read()
		if err != nil {
			panic(err)
		}

		log.Printf("%s", event)
	}
}
