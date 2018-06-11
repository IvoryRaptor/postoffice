package mqtt

import "github.com/IvoryRaptor/postoffice"

func (m *MQTT)Config(kernel postoffice.IPostOffice, config map[string]interface{}) error {
	m.kernel = kernel
	m.config = config
	return nil
}

func (m *MQTT)Start() error {

	return nil
}

func (m *MQTT)Stop(){

}
