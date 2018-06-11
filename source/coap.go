package source

import "github.com/IvoryRaptor/postoffice"

type CoapSource struct {
	kernel postoffice.IPostOffice
	ssl bool
}
func (s *CoapSource)Config(kernel postoffice.IPostOffice, config map[string]interface{}) error{
	return nil
}

func (s *CoapSource)Start() error{
	return nil
}

func (s *CoapSource)Stop(){

}
