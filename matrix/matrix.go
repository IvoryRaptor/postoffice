package matrix

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"fmt"
	"sync"
	"log"
	"encoding/base64"
)

type ZkMatrix struct {
	postoffice.Matrix
	secretChan    chan bool
	actionChan    chan bool
}

func (m *ZkMatrix) WatchSecret(kernel postoffice.IKernel,conn *zk.Conn) {
	go func() {
		for ; ; {
			secret, _, childCh, _ := conn.GetW(fmt.Sprintf("/postoffice/%s", m.Name))
			m.Authorization = base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s",m.Name,secret)))
			log.Printf("matrix %s: %s", m.Name, secret)
			select {
			case ev := <-childCh:
				if ev.Err != nil {
					fmt.Printf("Child watcher error: %+v\n", ev.Err)
					return
				}
			case <-m.secretChan:
				return
			}
		}
	}()
}

func (m *ZkMatrix) WatchAction(kernel postoffice.IKernel,conn *zk.Conn) {
	go func() {
		for ; ; {
			actions, _, childCh, _ := conn.ChildrenW(fmt.Sprintf("/postoffice/%s", m.Name))
			newAction := sync.Map{}
			for _, actionName := range actions {
				action, ok := m.Action.Load(actionName)
				if ok {
					newAction.Store(actionName, action)
					m.Action.Delete(actionName)
				} else {
					act := &Action{Name: actionName}
					act.WatchTopic(kernel, m.Name, conn)
					newAction.Store(actionName, action)
				}
				log.Printf("%s/%s", m.Name, actionName)
			}
			m.Action.Range(func(k, v interface{}) bool {
				v.(*Action).StopWatch()
				return true
			})
			m.Action = newAction
			select {
			case ev := <-childCh:
				if ev.Err != nil {
					fmt.Printf("Child watcher error: %+v\n", ev.Err)
					return
				}
			case <-m.actionChan:
				return
			}
		}
	}()
}

func (m *ZkMatrix) StopWatch() {
	m.secretChan <- true
	m.actionChan <- true
}
