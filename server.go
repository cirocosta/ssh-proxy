package main

import (
	"context"

	"github.com/cirocosta/wireguard-vs-ssh/ssh"
)

type serverCommand struct {
	Address string `long:"address" default:"0.0.0.0:2222" description:"address to list on"`
	Port    uint16 `long:"port"    default:"0"            description:"port to receive conns forward"`

	// PrivateKey can be generated with `ssh-keygen -t rsa`
	//
	PrivateKey string `long:"private-key" required:"true" description:"file location of a private key"`
}

func (c *serverCommand) Execute(args []string) (err error) {
	s, err := ssh.NewServer(c.Address, c.PrivateKey, c.Port)
	if err != nil {
		return
	}

	err = s.Start(context.Background())
	return
}
