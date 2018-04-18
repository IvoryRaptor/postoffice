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
)

type Mongo struct {
	kernel postoffice.IKernel
}

var GlobalMgoSession *mgo.Session

func init() {
	globalMgoSession, err := mgo.DialWithTimeout("mongodb://192.168.41.170:30707", 10 * time.Second)
	if err != nil {
		panic(err)
	}
	GlobalMgoSession=globalMgoSession
	GlobalMgoSession.SetMode(mgo.Monotonic, true)
	//default is 4096
	GlobalMgoSession.SetPoolLimit(300)
}

func CloneSession() *mgo.Session {
	return GlobalMgoSession.Clone()
}

func (a *Mongo) Config(kernel postoffice.IKernel,config *Config) error{
	return nil
}

func (a *Mongo) Start() error{
	return nil
}

func (a *Mongo) Authenticate(msg *message.ConnectMessage) *postoffice.ChannelConfig {
	sp := strings.Split(string(msg.Username()), "&")
	config := postoffice.ChannelConfig{
		DeviceName: sp[0],
		ProductKey: sp[1],
	}

	sp = strings.Split(string(msg.ClientId()), "|")
	if len(sp) != 3 {
		return nil
	}
	config.ClientId = sp[0]
	sp = strings.Split(sp[1], ",")
	params := map[string]string{
		"productKey": config.ProductKey,
		"deviceName": config.DeviceName,
		"clientId":   config.ClientId,
	}

	session := CloneSession() //调用这个获得session
	defer session.Close()     //一定要记得释放

	c := session.DB("tortoise").C("oauth_devices")
	data := bson.M{}
	err := c.Find(bson.M{"productKey": config.ProductKey, "deviceName": config.DeviceName}).One(data)
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
		h.Write([]byte(key))
		h.Write([]byte(params[key]))
	}
	if string(msg.Password()) != fmt.Sprintf("%X", h.Sum(nil)) {
		return nil
	}
	return &config
}
