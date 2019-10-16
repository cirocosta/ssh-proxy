package ssh

import (
	"context"
	"fmt"
	"net"

	"golang.org/x/crypto/ssh"
)

type client struct {
	// addr is the address of the SSH server (including the server's port).
	//
	addr string

	// port is the local port to be asked to be forwarded by the server.
	//
	port uint16

	client *ssh.Client
}

func NewClient(addr string, port uint16) client {
	return client{
		addr: addr,
		port: port,
	}
}

// Start starts a port-forwarding session, letting connections performed against
// a servers' port be forwarded to this client's port.
//
// ps.: this will block until either a failure occurs, or the context is
//      explicity cancelled.
//
func (c *client) Start(ctx context.Context) (err error) {
	sshClient, err := connect(ctx, c.addr)
	if err != nil {
		err = fmt.Errorf("failed to connect to ssh server: %w", err)
		return
	}

	defer sshClient.Close()

	// do the port-forwarding

	return
}

var defaultClientSSHConfig = &ssh.ClientConfig{
	Config: defaultSSHConfig,
	User:   "anything",
}

func connect(ctx context.Context, addr string) (sshClient *ssh.Client, err error) {
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		err = fmt.Errorf("failed to dial %s: %w", addr, err)
		return
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, defaultClientSSHConfig)
	if err != nil {
		err = fmt.Errorf("failed performing SSH handshake against ssh server: %w", err)
	}

	sshClient = ssh.NewClient(sshConn, chans, reqs)

	return
}
