package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"strings"
	"time"
	"gopkg.in/mgo.v2/bson"
	"crypto/sha1"
	"sort"
	"hash"
	"crypto/md5"
	"fmt"
	"crypto/hmac"
	"math/rand"
	"github.com/IvoryRaptor/dragonfly"
	"log"
)

type MongoAuth struct {
	dragonfly.Mongo
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (m *MongoAuth) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	clientId := string(msg.ClientId())
	deviceName := ""
	productKey := ""

	session := m.GetSession() //调用这个获得session
	defer session.Close()     //一定要记得释放
	c := session.DB("tortoise").C("oauth_devices")

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
		data := bson.M{}
		err := c.Find(bson.M{"matrix": productKey, "deviceName": deviceName}).One(data)
		if err != nil {
			log.Printf("Not found Matrix: %s DeviceName: %s", productKey, deviceName)
			return nil
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
					h = hmac.New(sha1.New, []byte(data["secret"].(string)))
				case "hmacmd5":
					h = hmac.New(md5.New, []byte(data["secret"].(string)))
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
		//time.Now().AddDate(0, 0, 1)
		//time.Date(2014, time.November, 5, 0, 0, 0, 0, time.UTC)

		if !strings.EqualFold(string(msg.Password()), fmt.Sprintf("%X", h.Sum(nil))) {
			return nil
		}
	} else {
		data := bson.M{}
		err := c.Find(bson.M{"token": clientId,"time":bson.M{"$gte":time.Now()}}).One(data)
		if err != nil {
			return nil
		}
		deviceName = data["deviceName"].(string)
		productKey = data["iotnn"].(string)
	}
	token := randSeq(8)
	c.Update(
		bson.M{"iotnn": productKey, "deviceName": deviceName},
		bson.M{"$set": bson.M{
			"token": token,
			"time":  time.Now().AddDate(0, 0, 2),
		}},
	)
	config := postoffice.ChannelConfig{
		DeviceName: deviceName,
		Matrix:     productKey,
		Token:      token,
	}
	return &config
}
