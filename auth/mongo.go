package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"strings"
	"gopkg.in/mgo.v2"
	"time"
	"gopkg.in/mgo.v2/bson"
	"crypto/sha1"
	"sort"
	"hash"
	"crypto/md5"
	"fmt"
	"crypto/hmac"
	"math/rand"
)

type Mongo struct {
	kernel postoffice.IKernel
}

var GlobalMgoSession *mgo.Session

func CloneSession() *mgo.Session {
	return GlobalMgoSession.Clone()
}

func (a *Mongo) Config(kernel postoffice.IKernel,config *Config) error{
	globalMgoSession, err := mgo.DialWithTimeout(config.Url, 10 * time.Second)
	if err != nil {
		panic(err)
	}
	GlobalMgoSession=globalMgoSession
	GlobalMgoSession.SetMode(mgo.Monotonic, true)
	//default is 4096
	GlobalMgoSession.SetPoolLimit(300)
	return nil
}

func (a *Mongo) Start() error{
	return nil
}

var letters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randSeq(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func (a *Mongo) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	clientId := string(msg.ClientId())
	deviceName := ""
	productKey := ""

	session := CloneSession() //调用这个获得session
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
		err := c.Find(bson.M{"productKey": productKey, "deviceName": deviceName}).One(data)
		if err != nil {
			println(err.Error())
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
					println(v[1])
				}
			case "timestamp":
				keys = append(keys, v[0])
				params[v[0]] = v[1]
			}
		}
		sort.Strings(keys)
		for _, key := range keys {
			println(key + "=" + params[key])
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
		productKey = data["productKey"].(string)
	}
	token := randSeq(8)
	c.Update(
		bson.M{"productKey": productKey, "deviceName": deviceName},
		bson.M{"$set": bson.M{
			"token": token,
			"time":  time.Now().AddDate(0, 0, 2),
		}},
	)
	config := postoffice.ChannelConfig{
		DeviceName: deviceName,
		ProductKey: productKey,
		Token:      token,
	}
	return &config
}
