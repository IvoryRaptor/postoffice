package matrix

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"fmt"
	"sync"
	"log"
)

type ZkMatrix struct {
	Name          string
	Action        sync.Map
	secretChan    chan bool
	actionChan    chan bool
}

type ZkAction struct {
	run       bool
	Name      string
	Topic    []string
	actionChan chan bool
}

func (m *ZkMatrix)GetTopics(action string) ([]string,bool) {
	r, ok := m.Action.Load(action)
	if !ok {
		return nil, ok
	}
	act := r.(*ZkAction)
	return act.Topic, ok
}

func (m *ZkMatrix) WatchSecret(kernel postoffice.IKernel,conn *zk.Conn) {
	go func() {
		for ; ; {
			secret, _, childCh, _ := conn.GetW(IOTNN_PATH + "/" + m.Name)
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
			actions, _, childCh, _ := conn.ChildrenW(IOTNN_PATH + "/" + m.Name)
			newAction := sync.Map{}
			for _, actionName := range actions {
				action, ok := m.Action.Load(actionName)
				if ok {
					newAction.Store(actionName, action)
					m.Action.Delete(actionName)
				} else {
					act := &ZkAction{}
					act.Name = actionName
					act.WatchTopic(kernel, m.Name, conn)
					newAction.Store(actionName, act)
				}
				log.Printf("%s/%s", m.Name, actionName)
			}
			m.Action.Range(func(k, v interface{}) bool {
				v.(*ZkAction).StopWatch()
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
