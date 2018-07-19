package auth

import (
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"crypto/hmac"
	"crypto/md5"
	"sort"
	"fmt"
	"gopkg.in/mgo.v2/bson"
	"crypto/sha1"
	"strings"
	"hash"
	"log"
)

type RedisAuth struct {
	dragonfly.Redis
}

func (m *RedisAuth) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	clientId := string(msg.ClientId())
	deviceName := ""
	productKey := ""

	if strings.Index(clientId, "|") > 0 {
		sp := strings.Split(string(msg.Username()), "&")
		if len(sp) != 2 {
			return nil
		}
		deviceName = sp[0]
		productKey = sp[1]

		sp = strings.Split(string(msg.ClientId()), "|")
		if len(sp) != 3 {
			return nil
		}
		clientId = sp[0]
		sp = strings.Split(sp[1], ",")
		params := map[string]string{
			"clientId":   clientId,
			"productKey": productKey,
			"deviceName": deviceName,
		}
		secret, err := m.Do("HGET", productKey, deviceName)
		if secret == nil || err != nil {
			fmt.Printf("Not found Matrix: %s DeviceName: %s", productKey, deviceName)
		}
		keys := []string{"productKey", "deviceName", "clientId"}
		var h hash.Hash = nil
		for _, i := range sp {
			v := strings.Split(i, "=")
			switch v[0] {
			case "securemode":
			case "signmethod":
				switch v[1] {
				case "hmacsha1":
					h = hmac.New(sha1.New, []byte(secret.(string)))
				case "hmacmd5":
					h = hmac.New(md5.New, []byte(secret.(string)))
				default:
					log.Printf("Unknown signmethod %s", v[1])
					return nil
				}
			case "timestamp":
				keys = append(keys, v[0])
				params[v[0]] = v[1]
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			h.Write([]byte(key))
			h.Write([]byte(params[key]))
		}
		if !strings.EqualFold(string(msg.Password()), fmt.Sprintf("%X", h.Sum(nil))) {
			return nil
		}
	} else {
		data := bson.M{}
		token, err := m.Do("GET", productKey+"@"+deviceName+"TOKEN")
		if token == nil || err != nil {
			fmt.Printf("Not Found Matrix: %s DeviceName: %s", productKey, deviceName)
			return nil
		}
		if token.(string) != clientId {
			return nil
		}
		deviceName = data["deviceName"].(string)
		productKey = data["iotnn"].(string)

	}
	token := randSeq(8)
	m.Do("SETEX", productKey+"@"+deviceName+"TOKEN", 60, token)
	config := postoffice.ChannelConfig{
		DeviceName: deviceName,
		Matrix:     productKey,
		Token:      token,
	}
	return &config
}
