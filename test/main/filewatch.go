package main

import (
	"github.com/fsnotify/fsnotify"
	"log"
	"os"
	"strings"
)

func ExampleNewWatcher(watchfile string) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	done := make(chan bool)
	go func() {
		for {
			select {
			case event := <-watcher.Events:
				log.Println("event:", event)
				if event.Op&fsnotify.Write == fsnotify.Write {
					log.Println("modified file:", event.Name)
				}
			case err := <-watcher.Errors:
				log.Println("error:", err)
			}
		}
	}()

	err = watcher.Add(watchfile)
	if err != nil {
		log.Fatal(err)
	}
	<-done
}

func main() {
	dir := "/tmp/aaa"
	if len(os.Args) >= 2 {
		dir = os.Args[1]
	}
	index := strings.LastIndex(dir, "/")
	println(dir[:index])
	//ExampleNewWatcher(dir)
}
