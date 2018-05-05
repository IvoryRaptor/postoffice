package source

import "github.com/IvoryRaptor/postoffice"

type CoapSource struct {
	kernel postoffice.IKernel
	ssl bool
}
func (s *CoapSource)Config(kernel postoffice.IKernel, config map[string]interface{}) error{
	return nil
}

func (s *CoapSource)Start() error{
	return nil
}

func (s *CoapSource)Stop(){

}
