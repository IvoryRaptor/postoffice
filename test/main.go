package main

import (
	"github.com/IvoryRaptor/postoffice-plus"
	"plugin"
	"regexp"
)

func TTT(conf map[interface{}]interface{}) postoffice_plus.IWorkPlus {
	p, err := plugin.Open("./other/p1.so")
	if err != nil {
		panic(err)
	}
	config, err := p.Lookup("Factory")
	if err != nil {
		panic(err)
	}

	res, err := config.(func(config map[interface{}]interface{}) (postoffice_plus.IWorkPlus, error))(conf)
	return res
}

var PlusRegexp = regexp.MustCompile(`^plus\.[\w]+`)

func main() {
	var d = "plus.reply"
	var d2 = "plus"

	//println("plus.reply"[0:5])
	println(d[0:5])

	if len(d2) > 5 && d2[0:5] == "plus." {
		println(d2[0:5])
	}

	//println("[" + PlusRegexp.FindString("plus.reply") + "]")
	//println("[" + PlusRegexp.FindString("plus_reply") + "]")
	//println("[" + PlusRegexp.FindString("asdf.") + "]")
	//
	//conf1 := map[interface{}]interface{}{
	//	"name": "test1",
	//}
	//conf2 := map[interface{}]interface{}{
	//	"name": "test2",
	//}
	//r1 := TTT(conf1)
	//r2 := TTT(conf2)
	//
	//r1.Work(&postoffice_plus.MQMessage{})
	//r2.Work(&postoffice_plus.MQMessage{})
}
