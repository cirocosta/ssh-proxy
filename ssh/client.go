package ssh

import (
	"context"
	"fmt"
	"net"
	"strconv"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// client is an SSH client that connects to an SSH server for the sole purpose
// of creating a port-forwarding session.
//
//
// 	user
//        |
// 	  *---- server:rport
//               |
//               *---> client
//                       |
//                       *---> application:lport
//
//
type client struct {
	// addr is the address of the SSH server (including the server's SSH port).
	//
	saddr string

	// lport is the port to which this client should create connections to.
	//
	lport uint16

	// rport is the remote port from which the server should get connectinos
	// to to forward to us.
	//
	rport uint16
}

var defaultClientSSHConfig = &ssh.ClientConfig{
	Config:          defaultSSHConfig,
	User:            "anything",
	HostKeyCallback: ssh.InsecureIgnoreHostKey(),
}

func NewClient(saddr string, lport, rport uint16) client {
	return client{saddr, lport, rport}
}

// Start starts a port-forwarding session, letting connections performed against
// a servers' port be forwarded to this client's port.
//
// ps.: this will block until either a failure occurs, or the context is
//      explicity cancelled.
//
func (c *client) Start(ctx context.Context) (err error) {
	sshClient, err := connect(ctx, c.saddr)
	if err != nil {
		err = fmt.Errorf("failed to connect to ssh server at %s: %w", c.saddr, err)
		return
	}

	defer sshClient.Close()

	log.Info("connected")

	err = c.portForward(ctx, sshClient)
	if err != nil {
		err = fmt.Errorf("failed port-forwarding: %w", err)
		return
	}

	return
}

// portForward asks the server on the other side of the connection to get
// connections forwarded to a given `port` on our side.
//
func (c *client) portForward(ctx context.Context, client *ssh.Client) (err error) {
	var (
		remoteAddr = "0.0.0.0:" + strconv.Itoa(int(c.rport))
		localAddr  = "127.0.0.1:" + strconv.Itoa(int(c.lport))
	)

	listener, err := client.Listen("tcp", remoteAddr)
	if err != nil {
		err = fmt.Errorf("failed requesting forward for %s: %w", remoteAddr, err)
		return
	}

	defer listener.Close()

	for {
		remoteConn, err := listener.Accept()
		if err != nil {
			return fmt.Errorf("failed accepting: %w", err)
		}

		err = handleForwardedConn(ctx, remoteConn, localAddr)
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
