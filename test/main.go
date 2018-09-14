package main

import (
	"encoding/json"
	"github.com/jmoiron/jsonq"
	"html/template"
	"os"
	"strings"
)

func main() {
	data := map[string]interface{}{}
	dec := json.NewDecoder(strings.NewReader(`{"foo": 1,
	"bar": 2,
	"test": "Hello, world!",
	"baz": 123.1,
	"array": [
		{"foo": 1},
		{"bar": 2},
		{"baz": 3}
	],
	"subobj": {
		"foo": 1,
		"subarray": [1,2,3],
		"subsubobj": {
			"bar": 2,
			"baz": 3,
			"array": ["hello", "world"]
		}
	},
	"bool": true}`))

	dec.Decode(&data)
	jq := jsonq.NewQuery(data)
	var d = []string{"subobj", "subsubobj", "array", "0"}

	s, err := jq.String(d...)
	if err != nil {
		println(err.Error())
	} else {
		println(s)
	}

	tmpl := template.New("tmpl1")
	_, err = tmpl.Parse(`Hello {{.String("subobj", "subsubobj", "array", "0")}} Welcome to go programming...\n`)
	if err != nil {
		println(err.Error())
		return
	}
	err = tmpl.Execute(os.Stdout, jq)
	if err != nil {
		println(err.Error())
	}
}
