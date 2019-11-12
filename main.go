package main

import (
	"log"
	"os"

	coreapp "github.com/ftl/panacotta/core/app"
	"github.com/ftl/panacotta/core/cfg"
	uiapp "github.com/ftl/panacotta/ui/app"
)

func main() {
	configuration, err := cfg.Load()
	if err != nil {
		log.Println(err)
		configuration = cfg.Static()
	}

	controller := coreapp.New(configuration)
	uiapp.Run(controller, os.Args)
}
