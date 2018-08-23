package auth

import (
	"crypto/hmac"
	"crypto/md5"
	"crypto/sha1"
	"fmt"
	"github.com/IvoryRaptor/dragonfly"
	"github.com/IvoryRaptor/postoffice"
	"hash"
	"log"
	"sort"
	"strings"
)

type RedisAuth struct {
	dragonfly.Redis
}

//func (r *RedisAuth) Stop() {
//	r.Conn.Close()
//}

func (m *RedisAuth) Authenticate(block *postoffice.AuthBlock) *postoffice.ChannelConfig {
	switch block.SecureMode {
	case 2:
		v, err := m.Do("HGET", block.ProductKey, block.DeviceName)
		if err != nil {
			fmt.Printf("Redis %s\n", err.Error())
			return nil
		}
		if v == nil {
			fmt.Printf("Not found Matrix: %s DeviceName: %s\n", block.ProductKey, block.DeviceName)
			return nil
		}
		var secret = []byte(v.([]uint8))
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
	case 99:
		v, err := m.Do("GET", block.ProductKey+"@"+block.Username)
		if err != nil {
			fmt.Printf("Redis %s\n", err.Error())
			return nil
		}
		if v == nil {
			fmt.Printf("Not found Matrix: %s DeviceName: %s\n", block.ProductKey, block.DeviceName)
			return nil
		}
		block.DeviceName = string([]byte(v.([]uint8)))
	default:
		log.Printf("Redis Auth Unknown securemode %d", block.SecureMode)
		return nil
	}
	token := randSeq(8)
	m.Do("SETEX", block.ProductKey+"@"+token, 60, block.DeviceName)
	config := postoffice.ChannelConfig{
		DeviceName: block.DeviceName,
		Matrix:     block.ProductKey,
		Token:      token,
	}
	return &config
}
