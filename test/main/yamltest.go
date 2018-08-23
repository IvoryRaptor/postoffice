package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

//type Route struct {
//	RouteR map[string][]string
//}
//
//type Matrix struct {
//	MatrixR map[string]Route
//}

func change() []string {

	data, err := ioutil.ReadFile("./config/iotnn/matrix.yaml")
	if err != nil {
		return nil
	}
	conf := map[string]map[string][]string{}
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		println("json error")
	}
	println("sss")
	return []string{""}
}

func change2() []string {

	data, err := ioutil.ReadFile("./config/auth_group/config.yaml")
	if err != nil {
		return nil
	}
	conf := make(map[string]interface{})
	err = yaml.Unmarshal(data, &conf)
	if err != nil {
		println("json error")
	}
	println("sss")
	return []string{""}
}
func main() {

	change2()
}
