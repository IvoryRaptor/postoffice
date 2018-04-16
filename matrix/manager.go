package matrix

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/samuel/go-zookeeper/zk"
	"sync"
	"fmt"
	"time"
	"log"
)

type Manager struct {
	kernel    postoffice.IKernel
	oauth     string
	zkHost    string
	conn      *zk.Conn
	matrixMap *sync.Map
}

func (m *Manager) Config(kernel postoffice.IKernel, config *Config) error {
	m.kernel = kernel
	m.oauth = config.OAuth
	m.zkHost = fmt.Sprintf("%s:%d",config.Zookeeper.Host,config.Zookeeper.Port)
	m.matrixMap = &sync.Map{}
	return nil
}

func (m *Manager) GetMatrix(name string) (*postoffice.Matrix, bool) {
	res, ok := m.matrixMap.Load(name)
	return res.(*postoffice.Matrix), ok
}

func (m *Manager) Start() error {
	var err error
	m.conn, _, err = zk.Connect([]string{m.zkHost}, time.Second*10)
	if err != nil {
		return err
	}
	go func() {
		for ; ; {
			newMap := sync.Map{}
			keys, _, childCh, _ := m.conn.ChildrenW("/postoffice")
			for _, key := range keys {
				matrix, ok := m.matrixMap.Load(key)
				if ok {
					newMap.Store(key, matrix)
					newMap.Delete(key)
				} else {
					matrix := &ZkMatrix{}
					matrix.Name = key
					matrix.WatchSecret(m.kernel, m.conn)
					matrix.WatchAction(m.kernel, m.conn)
					newMap.Store(key, matrix)
				}
			}
			m.matrixMap.Range(func(k, v interface{}) bool {
				v.(*ZkMatrix).StopWatch()
				return true
			})
			m.matrixMap = &newMap
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
