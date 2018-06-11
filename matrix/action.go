package matrix

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"fmt"
	"log"
)

const IOTNN_PATH="/iotnn"

func (a *ZkAction) WatchTopic(kernel postoffice.IPostOffice, matrix string,conn *zk.Conn) {
	a.run = true
	go func() {
		for ; ; {
			t, _, childCh, _ := conn.ChildrenW(IOTNN_PATH + "/" + matrix +  "/" + a.Name)
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

func (a *ZkAction) StopWatch() {
	a.actionChan <- true
}
