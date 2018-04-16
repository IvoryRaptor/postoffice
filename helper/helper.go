package helper

var Base62 *Encoding
var Base36 *Encoding

func init() {
	Base62, _ = NewEncoding("0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")
	Base36, _ = NewEncoding("0123456789abcdefghijklmnopqrstuvwxyz")
}
