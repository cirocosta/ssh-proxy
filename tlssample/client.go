package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io"
	"log"
	"os"
)

var (
	addr = flag.String("addr", "", "address to connect to")
)

func client() (err error) {
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(cert)

	conf := &tls.Config{
		InsecureSkipVerify: true,
	}

	conn, err := tls.Dial("tcp", *addr, conf)
	if err != nil {
		return
	}
	defer conn.Close()

	log.Println("connected")

	f, err := os.Open("/dev/zero")
	if err != nil {
		return
	}
	defer f.Close()

	_, err = io.Copy(conn, f)
	return
}
