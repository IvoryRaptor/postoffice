package iotnn

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"fmt"
	"time"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice2/iotnn"
)

type ZkIotNN struct {
	kernel   postoffice.IPostOffice
	url      string
	conn     *zk.Conn
	iotnnMap *sync.Map
}

type IIotNN interface {
	dragonfly.IService
	GetMatrix(matrix string) (iotnn.IMatrix,bool)
}

func (m *ZkIotNN) Config(kernel dragonfly.IKernel, config map[interface {}]interface{}) error {
	m.kernel = kernel.(postoffice.IPostOffice)
	m.url = fmt.Sprintf("%s:%d", config["host"].(string), config["port"].(int))
	m.iotnnMap = &sync.Map{}
	return nil
}

func (m *ZkIotNN) GetMatrix(name string) (iotnn.IMatrix, bool) {
	res, ok := m.iotnnMap.Load(name)
	if ok{
		return res.(*ZkMatrix), ok
	}
	return nil,ok
}

func (m *ZkIotNN) Start() error {
	var err error
	m.conn, _, err = zk.Connect([]string{m.url}, time.Second*10)
	if err != nil {
		return err
	}

	flags := int32(0)
	acl := zk.WorldACL(zk.PermAll)
	m.conn.Create("/matrixs", []byte{}, flags, acl)
	m.conn.Create("/matrixs/default", []byte{}, flags, acl)
	m.conn.Create(fmt.Sprintf("/matrixs/default/postoffice-%s", m.kernel.Get("host")), []byte{}, flags, acl)

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
					iotnn := &ZkMatrix{}
					iotnn.Name = key
					iotnn.WatchSecret(m.kernel, m.conn)
					iotnn.WatchAction(m.kernel, m.conn)
					newMap.Store(key, iotnn)
				}
			}
			m.iotnnMap.Range(func(k, v interface{}) bool {
				v.(*ZkMatrix).StopWatch()
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

func (m *ZkIotNN) Stop() {
	m.kernel.RemoveService(m)
}
