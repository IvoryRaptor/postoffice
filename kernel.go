package postoffice

import (
	"net"
	"sync"
)

type Matrix struct {
	Name          string
	Authorization string
	Action        sync.Map
}

type IKernel interface {
	GetHost() int
	Start() error
	AddChannel(c net.Conn) (err error)
	GetMatrix(name string) (*Matrix, bool)
}
