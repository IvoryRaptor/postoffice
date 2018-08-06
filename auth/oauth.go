package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/dragonfly"
	"regexp"
	"log"
	"fmt"
	"io/ioutil"
	"net/http"
)

type IParser interface {
	Init(config map[interface{}]interface{}) error
	Parse(string) (string, bool)
}

type JsonParse struct {
}

func (p *JsonParse) Init(config map[interface{}]interface{}) error {
	return nil
}

func (p *JsonParse) Parse(string) (string, bool) {
	return "", false
}

type RegexpParser struct {
	regDevice *regexp.Regexp
}

func (p *RegexpParser) Init(config map[interface{}]interface{}) error {
	return nil
}

func (p *RegexpParser) Parse(string) (string, bool) {
	return "", false
}

//res := m.regDevice.FindStringSubmatch(string(body))

type OAuth struct {
	httpFmt string
	parser  IParser
}

func (m *OAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	switch block.SecureMode {
	case 98:
		url := fmt.Sprintf(m.httpFmt, block.Password)
		resp, err := http.Get(url)
		if err != nil {
			return nil
		}
		defer resp.Body.Close()
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Http %s :%s", url, err.Error())
			return nil
		}

		if len(res) == 0 {
			log.Printf("Http %s :%s", url, body)
			return nil
		}
		config := postoffice.ChannelConfig{
			DeviceName: res[0],
			Matrix:     block.ProductKey,
		}
		return &config
	default:
		log.Printf("Redis Auth Unknown securemode %d\n", block.SecureMode)
		return nil
	}
	return nil
}

func (m *OAuth) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
	m.httpFmt = config["Http"].(string)
	t := config["type"].(string)
	switch t {
	case "json":
		m.parser = &JsonParse{}
	case "regexp":
		m.parser = &RegexpParser{}
	default:
		log.Printf("\n", )
	}
	if m.parser != nil {
		m.parser.Init(config)
	}
	return nil
}

func (m *OAuth) Start() error {
	return nil
}

func (m *OAuth) Stop() {

}
