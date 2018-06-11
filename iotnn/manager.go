package iotnn

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"fmt"
	"time"
	"log"
)

type Manager struct {
	kernel   postoffice.IPostOffice
	oauth    string
	zkHost   string
	conn     *zk.Conn
	iotnnMap *sync.Map
}

type IMatrix interface {
	GetTopics(action string) ([]string,bool)
}

func (m *Manager) Config(kernel postoffice.IPostOffice, config *Config) error {
	m.kernel = kernel
	m.oauth = config.OAuth
	m.zkHost = fmt.Sprintf("%s:%d",config.Zookeeper.Host,config.Zookeeper.Port)
	println(m.zkHost)
	m.iotnnMap = &sync.Map{}
	return nil
}

func (m *Manager) GetMatrix(name string) (IMatrix, bool) {
	res, ok := m.iotnnMap.Load(name)
	if ok{
		return res.(*ZkIOTNN), ok
	}
	return nil,ok
}

func (m *Manager) Start() error {
	var err error
	m.conn, _, err = zk.Connect([]string{m.zkHost}, time.Second*10)
	if err != nil {
		return err
	}

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	m.conn.Create("/matrixs", []byte{}, flags, acl)
	m.conn.Create("/matrixs/default", []byte{}, flags, acl)
	m.conn.Create(fmt.Sprintf("/matrixs/default/postoffice-%s", m.kernel.GetHost()), []byte{}, flags, acl)

	go func() {
		for ; ; {
			newMap := sync.Map{}
			keys, _, childCh, _ := m.conn.ChildrenW(IOTNN_PATH)
			for _, key := range keys {
				iotnn, ok := m.iotnnMap.Load(key)
				if ok {
					newMap.Store(key, iotnn)
					m.iotnnMap.Delete(key)
				} else {
					iotnn := &ZkIOTNN{}
					iotnn.Name = key
					iotnn.WatchSecret(m.kernel, m.conn)
					iotnn.WatchAction(m.kernel, m.conn)
					newMap.Store(key, iotnn)
				}
			}
			m.iotnnMap.Range(func(k, v interface{}) bool {
				v.(*ZkIOTNN).StopWatch()
				return true
			})
			m.iotnnMap = &newMap
			select {
			case ev := <-childCh:
				if ev.Err != nil {
					fmt.Printf("Child watcher error: %+v\n", ev.Err)
					return
				}
			}
		}
	}()
	return nil
}

func (m *Manager) Stop() error {
	log.Println("router stop")
	return nil
}
