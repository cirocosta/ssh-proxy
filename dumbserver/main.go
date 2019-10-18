package main

import (
	"flag"
	"fmt"
	"net/http"
)

var (
	addr = flag.String("addr", ":8000", "address to bind the server to")
)

func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello, %s!", r.URL.Path[1:])
}

func main() {
	flag.Parse()

	http.HandleFunc("/", sayHello)

	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		panic(err)
	}
}
