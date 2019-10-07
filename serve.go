package main

import (
	"github.com/cirocosta/wireguard-vs-ssh/ssh"
)

type serveCommand struct {
	Address string `long:"address" default:"0.0.0.0:2222"`
}

func (c *serveCommand) Execute(args []string) (err error) {
	s := ssh.NewServer(c.Address)

	s.Serve()

	return
}
