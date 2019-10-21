package main

import (
	"flag"
)

var (
	isServer = flag.Bool("server", false, "whether it's a server or not")
)

func main() {
	flag.Parse()

	var fn = client

	if *isServer {
		fn = server
	}

	must(fn())
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
