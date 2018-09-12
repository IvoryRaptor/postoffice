package main

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/IvoryRaptor/postoffice-plus"
	"github.com/IvoryRaptor/postoffice-plus/mqtt/message"
	"io/ioutil"
	"log"
)

type Base64Writer struct {
	BaseUrl  string
	BasePath string
}

func (b *Base64Writer) Work(msg *postoffice_plus.MQMessage) ([]message.Message, error) {
	pus := message.NewPublishMessage()
	var payload map[string]interface{}
	err := json.Unmarshal(msg.Payload, &payload)
	data, err := base64.StdEncoding.DecodeString(payload["base64"].(string))
	if err != nil {
		log.Fatalln(err)
	}
	h := sha1.New()
	h.Write(data)
	bs := h.Sum(nil)
	filename := base64.StdEncoding.EncodeToString(bs)
	fullName := fmt.Sprintf("%s/%s.%s", b.BasePath, filename, payload["ext"].(string))
	fullUrl := fmt.Sprintf("%s/%s.%s", b.BaseUrl, filename, payload["ext"].(string))
	ioutil.WriteFile(fullName, data, 0644)

	topic := fmt.Sprintf(
		"%s/%s/%s/%s",
		msg.Matrix,
		msg.Device,
		msg.Resource,
		msg.Action)
	pus.SetTopic([]byte(topic))
	data, err = json.Marshal(map[string]interface{}{
		"url": fullUrl,
	})
	if err != nil {
		return nil, err
	}
	pus.SetPayload(data)
	return []message.Message{pus}, nil
}

func Factory(config map[interface{}]interface{}) (postoffice_plus.IWorkPlus, error) {
	r := Base64Writer{
		BaseUrl:  config["BaseUrl"].(string),
		BasePath: config["BasePath"].(string),
	}
	return &r, nil
}
