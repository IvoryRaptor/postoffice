package auth

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"github.com/surge/glog"
	"io/ioutil"
	"log"
	"net/http"
	"regexp"
	"strings"
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
	method  string
	headers map[interface{}]interface{}
	userid  string
	parser  IParser
}

// 微信公众号换取方式
// https://api.weixin.qq.com/sns/oauth2/access_token?appid=APPID&secret=SECRET&code=CODE&grant_type=authorization_code
//{ "access_token":"ACCESS_TOKEN",
//"expires_in":7200,
//"refresh_token":"REFRESH_TOKEN",
//"openid":"OPENID",
//"scope":"SCOPE" }

// 小程序换取方式 https://api.weixin.qq.com/sns/jscode2session?appid=APPID&secret=SECRET&js_code=JSCODE&grant_type=authorization_code
////正常返回的JSON数据包
//{
//"openid": "OPENID",
//"session_key": "SESSIONKEY",
//}
//
////满足UnionID返回条件时，返回的JSON数据包
//{
//"openid": "OPENID",
//"session_key": "SESSIONKEY",
//"unionid": "UNIONID"
//}
////错误时返回JSON数据包(示例为Code无效)
//{
//"errcode": 40029,
//"errmsg": "invalid code"
//}
//自定义换取方式
// curl --compressed -v localhost:8080/v1/oauth/tokens \
//-u test_client_1:test_secret \
//-d "grant_type=authorization_code" \
//-d "code=7afb1c55-76e4-4c76-adb7-9d657cb47a27" \
//-d "redirect_uri=https://www.example.com"
//{
//"user_id": "1",
//"access_token": "00ccd40e-72ca-4e79-a4b6-67c95e2e3f1c",
//"expires_in": 3600,
//"token_type": "Bearer",
//"scope": "read_write",
//"refresh_token": "6fd8d272-375a-4d8a-8d0f-43367dc8b791"
//}

func (m *OAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	switch block.SecureMode {
	case 98:
		url := fmt.Sprintf(m.httpFmt, block.DeviceName)
		client := &http.Client{}
		//fmt.Println(url)
		if strings.LastIndex(url, "https://") >= 0 {
			client.Transport = &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}
		}
		req, err := http.NewRequest(strings.ToUpper(m.method), url, nil)
		for k, v := range m.headers {
			req.Header.Add(k.(string), v.(string))
		}
		resp, err := client.Do(req)
		defer resp.Body.Close()
		if err != nil {
			return nil
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("Http %s :%s", url, err.Error())
			return nil
		}
		//println(body)
		////if len(res) == 0 {
		//println(len(body))
		//if len("") == 0 {
		//	log.Printf("Http %s :%s", url, body)
		//	return nil
		//}
		var response map[string]interface{}
		if err = json.Unmarshal(body, &response); err != nil {
			log.Println("json error")
			return nil
		}
		//response := f.(map[string]interface{})
		if _, ok2 := response[m.userid]; !ok2 {
			glog.Error("error code ", block.DeviceName)
			return nil
		}

		config := postoffice.ChannelConfig{
			DeviceName: response[m.userid].(string),
			Matrix:     block.ProductKey,
		}
		log.Println("config ok ")
		return &config
	default:
		log.Printf("Oauth Auth Unknown securemode %d\n", block.SecureMode)
		return nil
	}
	return nil
}

func (m *OAuth) Config(kernel dragonfly.IKernel, config map[interface{}]interface{}) error {
	m.httpFmt = config["http"].(string)
	t := config["type"].(string)
	switch t {
	case "json":
		m.parser = &JsonParse{}
	case "regexp":
		m.parser = &RegexpParser{}
	default:
		log.Printf("\n")
	}
	if m.parser != nil {
		m.parser.Init(config)
	}
	m.method = config["method"].(string)
	m.userid = config["userid"].(string)
	m.headers = config["headers"].(map[interface{}]interface{})
	return nil
}

func (m *OAuth) Start() error {
	return nil
}

func (m *OAuth) Stop() {

}
