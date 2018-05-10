package mqtt

import "github.com/IvoryRaptor/postoffice"

func (m *MQTT)Config(kernel postoffice.IKernel, config map[string]interface{}) error {
	m.kernel = kernel
	print("*********", kernel)
	m.config = config
	return nil
}

func (m *MQTT)Start() error {

	return nil
}

func (m *MQTT)Stop(){

}
