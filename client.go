package main

import (
	"context"

	"github.com/cirocosta/ssh-proxy/ssh"
)

type clientCommand struct {
	Address    string `long:"addr"        required:"true" description:"address to connect to"`
	RemotePort uint16 `long:"remote-port" required:"true" description:"port to be port-forwarded by the server"`
	LocalPort  uint16 `long:"local-port"  required:"true" description:"port to be port-forwarded by the server"`
}

func (c *clientCommand) Execute(args []string) (err error) {
	client := ssh.NewClient(c.Address, c.LocalPort, c.RemotePort)

	err = client.Start(context.Background())
	return
}
