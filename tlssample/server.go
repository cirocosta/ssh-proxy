package main

import (
	"crypto/tls"
	"flag"
	"io"
	"log"
	"net"
	"os"
)

var (
	port = flag.String("port", "443", "port to listen on")
)

func server() (err error) {
	cer, err := tls.X509KeyPair(cert, key)
	if err != nil {
		return
	}

	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	ln, err := tls.Listen("tcp", ":"+*port, config)
	if err != nil {
		return
	}

	defer ln.Close()

	log.Println("listening")

	var conn net.Conn

	for {
		conn, err = ln.Accept()
		if err != nil {
			return
		}

		log.Println("conn accepted")

		err = handleConnection(conn)
		if err != nil {
			return
		}
	}
}

func handleConnection(conn net.Conn) (err error) {
	defer conn.Close()

	f, err := os.OpenFile("/dev/null", os.O_RDWR, 0666)
	if err != nil {
		return
	}

	defer f.Close()

	_, err = io.Copy(f, conn)
	return
}
