package ssh

import (
	"context"
	"fmt"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
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

var defaultClientSSHConfig = &ssh.ClientConfig{
	Config:          defaultSSHConfig,
	User:            "anything",
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
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

	log.Info("connected")

	err = portForward(ctx, sshClient, c.port)
	if err != nil {
		err = fmt.Errorf("failed port-forwarding for %d: %w", c.port, err)
		return
	}

	return
}

// portForward asks the server on the other side of the connection to get
// connections forwarded to a given `port` on our side.
//
func portForward(ctx context.Context, client *ssh.Client, port uint16) (err error) {
	addr := "0.0.0.0:" + strconv.Itoa(int(port))

	listener, err := client.Listen("tcp", addr)
	if err != nil {
		err = fmt.Errorf("failed requesting forward for %d: %w", port, err)
		return
	}

	defer listener.Close()

	for {
		remoteConn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed accepting: %w", err)
		}

		err = handleForwardedConn(ctx, remoteConn, addr)
		if err != nil {
			return fmt.Errorf("failed handling forwarded conn: %w", err)
		}
	}

	return
}

func handleForwardedConn(ctx context.Context, remoteConn net.Conn, addr string) (err error) {
	defer remoteConn.Close()

	localConn, err := net.Dial("tcp", addr)
	if err != nil {
		err = fmt.Errorf("failed to dial local server %s: %w", err)
		return
	}

	handleTraffic(ctx, localConn, remoteConn)

	return
}

func connect(ctx context.Context, addr string) (client *ssh.Client, err error) {
	dialer := &net.Dialer{}

	conn, err := dialer.DialContext(ctx, "tcp", addr)
	if err != nil {
		err = fmt.Errorf("failed to dial %s: %w", addr, err)
		return
	}

	sshConn, chans, reqs, err := ssh.NewClientConn(conn, addr, defaultClientSSHConfig)
	if err != nil {
		err = fmt.Errorf("failed performing SSH handshake against ssh server: %w", err)
		return
	}

	client = ssh.NewClient(sshConn, chans, reqs)

	return
}
