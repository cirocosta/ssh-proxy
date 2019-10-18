package main

import (
	"context"

	"github.com/cirocosta/ssh-proxy/ssh"
)

type serverCommand struct {
	Address string `long:"addr" default:"0.0.0.0:2222" description:"address to bind to"`

	// PrivateKey can be generated with `ssh-keygen -t rsa`
	//
	PrivateKey string `long:"private-key" required:"true" description:"file location of a private key"`
}

func (c *serverCommand) Execute(args []string) (err error) {
	s, err := ssh.NewServer(c.Address, c.PrivateKey)
	if err != nil {
		return
	}

	err = s.Start(context.Background())
	return
}
