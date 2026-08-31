package main

import (
	"github.com/tmpim/casket/casket/casketmain"
	_ "github.com/tmpim/casket/caskethttp/forwardproxy"
)

func main() {
	casketmain.Run()
}
