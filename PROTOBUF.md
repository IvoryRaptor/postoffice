export PATH=$PATH:$GOPATH/bin
protoc --go_out=. *.proto
