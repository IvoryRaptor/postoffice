package main

import (
	"github.com/IvoryRaptor/postoffice/kernel"
	"log"
	"os"
	"flag"
)


func main() {
	k := kernel.Kernel{
		ConfigFile: "./config/postoffice/config.yaml",
	}
	hostname := flag.String("hostname", os.Getenv("hostname"), "is ok")
	flag.Parse()
	println(*hostname)
	err := k.Config(*hostname)
	if err != nil {
		log.Fatalf(err.Error())
	}
	err = k.Start()
	if err != nil {
		log.Fatalf(err.Error())
	}
	k.WaitStop()
}
