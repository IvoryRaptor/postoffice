package main

import (
	"github.com/IvoryRaptor/postoffice/kernel"
	"log"
)


func main() {
	k := kernel.Kernel{
		ConfigFile: "./postoffice.yaml",
	}
	err := k.Config()
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = k.Start()
	if err != nil {
		log.Fatalf(err.Error())
	}
	k.WaitStop()
}
