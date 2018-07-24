package main

import (
	"flag"
	"os"
	"log"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice/kernel"
	"github.com/IvoryRaptor/postoffice/source"
	"github.com/IvoryRaptor/postoffice/auth"
	"github.com/IvoryRaptor/postoffice/mq"
)

func main() {
	k := kernel.PostOffice{}
	k.New("postoffice")

	hostname := flag.String("hostname", os.Getenv("hostname"), "is ok")
	flag.Parse()

	k.Set("matrix", "default")
	k.Set("angler", *hostname)

	err := dragonfly.Builder(
		&k,
		[]dragonfly.IServiceFactory{
			&source.Factory{},
			&auth.Factory{},
			&mq.Factory{},
			&dragonfly.RedisFactory{},
			&dragonfly.ZookeeperFactory{},
		})
	k.SetFields()
	if err != nil {
		log.Fatal(err.Error())
	}
	err = k.Start()
	if err != nil {
		log.Fatal(err.Error())
	}
	k.WaitStop()
}
