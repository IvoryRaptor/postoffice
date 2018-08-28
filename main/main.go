package main

import (
	"flag"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/iotnn"
	"github.com/IvoryRaptor/postoffice/kernel"
	"github.com/IvoryRaptor/postoffice/mq"
	"github.com/IvoryRaptor/postoffice/source"
	"log"
	"os"
)

func main() {
	k := kernel.PostOffice{}
	k.New("postoffice", k.SetFields)

	hostname := flag.String("hostname", os.Getenv("hostname"), "is ok")
	flag.Parse()

	k.Set("matrix", "default")
	k.Set("angler", *hostname)

	err := dragonfly.Builder(
		&k,
		[]dragonfly.IServiceFactory{
			&source.Singleton,
			&auth.Singleton,
			&mq.Singleton,
			&iotnn.Singleton,
			&dragonfly.Singleton,
		})
	err = k.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.WaitStop()
}
