package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"time"
	"gopkg.in/mgo.v2/bson"
	"math/rand"
	"github.com/IvoryRaptor/dragonfly"
	"log"
	"fmt"
	"crypto/hmac"
	"crypto/sha1"
	"crypto/md5"
	"sort"
	"hash"
	"strings"
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

func (m *MongoAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	session := m.GetSession() //调用这个获得session
	defer session.Close()     //一定要记得释放
	c := session.DB("tortoise").C("oauth_devices")
	data := bson.M{}
	var secret []byte
	switch block.SecureMode {
	case 2:
		err := c.Find(bson.M{"matrix": block.ProductKey, "deviceName": block.DeviceName}).One(data)
		if err != nil {
			log.Printf("Not found Matrix: %s DeviceName: %s", block.ProductKey, block.DeviceName)
			return nil
		}
		secret = []byte(data["secret"].(string))
	case 99:
		err := c.Find(bson.M{"matrix": block.ProductKey, "deviceName": block.DeviceName}).One(data)
		if err != nil {
			log.Printf("Not found Matrix: %s DeviceName: %s", block.ProductKey, block.DeviceName)
			return nil
		}
		secret = []byte(data["token"].(string))
	default:
		log.Printf("Redis Auth Unknown securemode %d", block.SecureMode)
		return nil
	}

	var h hash.Hash = nil
	switch block.SignMethod {
	case "hmacsha1":
		h = hmac.New(sha1.New, secret)
	case "hmacmd5":
		h = hmac.New(md5.New, secret)
	default:
		log.Printf("Unknown signmethod %s", block.SignMethod)
		return nil
	}
	sort.Strings(block.Keys)
	for _, key := range block.Keys {
		h.Write([]byte(key))
		h.Write([]byte(block.Params[key]))
	}
	if !strings.EqualFold(block.Password, fmt.Sprintf("%X", h.Sum(nil))) {
		return nil
	}
	token := randSeq(8)
	c.Update(
		bson.M{"iotnn": block.ProductKey, "deviceName": block.DeviceName},
		bson.M{"$set": bson.M{
			"token": token,
			"time":  time.Now().AddDate(0, 0, 2),
		}},
	)
	config := postoffice.ChannelConfig{
		DeviceName: block.DeviceName,
		Matrix:     block.ProductKey,
		Token:      token,
	}
	return &config
}
