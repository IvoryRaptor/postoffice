package auth

import (
	"github.com/IvoryRaptor/postoffice"
	"github.com/IvoryRaptor/postoffice/mqtt/message"
	"github.com/IvoryRaptor/postoffice/helper"
	"fmt"
	"net/http"
	"io/ioutil"
)

type OAuth struct {
	url    string
	kernel postoffice.IKernel
}

func (a *OAuth) Config(kernel postoffice.IKernel,config *Config) error{
	a.url = config.Url
	return nil
}

func (a *OAuth) Start() error{
	return nil
}

func (a *OAuth) Authenticate(msg *message.ConnectMessage) error {
	username := string(msg.Username())
	data, err := helper.Base62.Decode(username)
	if err != nil {
		return err
	}
	if len(data) != 32 {
		return ErrAuthFailure
	}

	matrix := helper.Base36.Encode(data[0:15])
	action := fmt.Sprintf("%x", data[16:])

	m, ok := a.kernel.GetMatrix(matrix)
	if !ok{
		return ErrAuthFailure
	}

	fmt.Println(matrix, action)

	url:=fmt.Sprintf("%s?grant_type=authorization_code&code=%s", a.url, msg.Password())

	req, err := http.NewRequest("GET", url, nil)

	req.Header.Set("Authorization", m.Authorization)
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println("response Body:", string(body))

	fmt.Println(string(body))
	//http://localhost:8080/v1/oauth/tokens
	//-u test_client_1:test_secret \
	//-d "grant_type=authorization_code" \
	//-d "code=7afb1c55-76e4-4c76-adb7-9d657cb47a27" \
	//-d "redirect_uri=https://www.example.com"
	//matrix action
	return nil
}
