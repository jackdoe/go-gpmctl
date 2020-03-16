package main

import (
	"log"

	gpmctl "github.com/jackdoe/go-gpmctl"
)

func main() {
	g, err := gpmctl.NewGPM(gpmctl.DefaultConf)
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
