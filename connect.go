package main

import (
	"context"

	"github.com/cirocosta/wireguard-vs-ssh/ssh"
)

type connectCommand struct {
	Address string `long:"address" required:"true" description:"address to connect to"`
	Port    uint16 `long:"port"    required:"true" description:"port to be port-forwarded by the server"`
}

func (c *connectCommand) Execute(args []string) (err error) {
	client := ssh.NewClient(c.Address)

	err = client.Start(context.Background())
	return
}
