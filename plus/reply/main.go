package main

import (
	"fmt"
	"github.com/IvoryRaptor/postoffice-plus"
	"github.com/IvoryRaptor/postoffice-plus/message"
)

type Reply struct {
}

func (r *Reply) Work(msg *postoffice_plus.MQMessage) ([]message.Message, error) {
	pus := message.NewPublishMessage()
	topic := fmt.Sprintf(
		"%s/%s/%s/%s",
		msg.Matrix,
		msg.Device,
		msg.Resource,
		msg.Action)
	pus.SetTopic([]byte(topic))
	pus.SetPayload(msg.Payload)
	return []message.Message{pus}, nil
}

func Factory(config map[interface{}]interface{}) (postoffice_plus.IWorkPlus, error) {
	return &Reply{}, nil
}
