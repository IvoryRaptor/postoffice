package matrix

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"fmt"
	"log"
)

type Action struct {
	run       bool
	Name      string
	Topic    []string
	actionChan chan bool
}


func (a *Action) WatchTopic(kernel postoffice.IKernel, matrix string,conn *zk.Conn) {
	a.run = true
	go func() {
		for ; ; {
			t, _, childCh, _ := conn.ChildrenW(fmt.Sprintf("/postoffice/%s/%s", matrix, a.Name))
			a.Topic = t
			log.Printf("%s %s -> %s", matrix, a.Name, t)
			select {
			case ev := <-childCh:
				if ev.Err != nil {
					fmt.Printf("Child watcher error: %+v\n", ev.Err)
					return
				}
			case <-a.actionChan:
				return
			}
		}
	}()
}

func (a *Action) StopWatch() {
	a.actionChan <- true
}
