package main

import (
	"flag"
	"github.com/IvoryRaptor/postoffice/kernel"
	"log"
	"os"
)

func main() {
	k := kernel.PostOffice{}
	hostname := flag.String("hostname", os.Getenv("hostname"), "is ok")
	flag.Parse()
	err := k.New(*hostname)
	if err != nil {
		log.Fatal(err.Error())
	}
	err = k.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.WaitStop()
}
