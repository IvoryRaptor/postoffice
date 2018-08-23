package main

func main() {
	a := map[string]string{"name": "aa"}
	if a["name"] == "" {
		println("aaa")
	} else {
		println(a["name"])
	}

	b := map[string]int{"aa": 10}
	if v, ok := b["dd"]; ok {

	}
}
