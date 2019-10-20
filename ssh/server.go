package ssh

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strconv"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

type (

	// server is a single-tenant SSH server that is capable of just one
	// thing: port-forwarding.
	//
	server struct {
		address string
		config  *ssh.ServerConfig
	}

	// tcpipForwardRequest is the request sent by the SSH client to request the
	// server to start a proxy that is meant to forward requests to a given port.
	//
	tcpipForwardRequest struct {

		// BindIP is the IP from the client that we, as the server, should reach
		// out to.
		//
		BindIP string

		// BindPort is the port in the client where we should direct our
		// requests to when proxying.
		//
		BindPort uint32
	}

	// tcpipForwardResponse represents the response sent back to the client who
	// asked for port-forwarding.
	//
	tcpipForwardResponse struct {

		// BoundPort is the port that was created on the server-side to bind to
		// the client's port.
		//
		BoundPort uint32
	}

	// forwardTCPIPChannelRequest TODO
	//
	forwardTCPIPChannelRequest struct {
		ForwardIP   string
		ForwardPort uint32
		OriginIP    string
		OriginPort  uint32
	}
)

func NewServer(address, pkey string) (s server, err error) {
	config, err := sshServerConfig(pkey)
	if err != nil {
		err = fmt.Errorf("failed generating ssh server config: %w", err)
		return
	}

	s = server{
		address: address,
		config:  config,
	}

	return
}

// Start serves an SSH server capable of port-forwarding.
//
func (server *server) Start(ctx context.Context) (err error) {

	// listen on a given address
	//
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		err = fmt.Errorf("failed to listen on %s: %w", server.address, err)
		return
	}

	log.WithFields(log.Fields{
		"addr": server.address,
	}).Info("listening")

	var tcpConn net.Conn

	for {
		// wait for someone to go through the tcp connection flow
		//
		tcpConn, err = listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error("failed to accept", err)
			}

			continue
		}

		// handle the application-layer (SSH)
		//
		err = server.handleSSH(ctx, tcpConn)
		if err != nil {
			err = fmt.Errorf("failed during SSH handling: %w", err)
			return
		}
	}
}

// sshServerConfig creates an SSH server configuration with a private key
// already loaded.
//
func sshServerConfig(filepath string) (cfg *ssh.ServerConfig, err error) {
	privateBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		err = fmt.Errorf("failed to read private key from %s: %w", err)
		return
	}

	signer, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		err = fmt.Errorf("failed to parse private key from %s: %w", err)
		return
	}

	cfg = &ssh.ServerConfig{
		Config:       defaultSSHConfig,
		NoClientAuth: true,
	}

	cfg.AddHostKey(signer)

	return
}

// handleSSH establishes the application-level (SSH) connection.
//
func (server *server) handleSSH(ctx context.Context, netConn net.Conn) (err error) {
	logger := log.WithFields(log.Fields{
		"remote": netConn.RemoteAddr().String(),
	})

	logger.Info("handling ssh conn")

	conn, chans, reqs, err := ssh.NewServerConn(netConn, server.config)
	if err != nil {
		err = fmt.Errorf("handshake failed: %w", err)
		return
	}

	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	go discardChannels(chans)

	err = server.handleSSHRequests(ctx, conn, reqs)
	if err != nil {
		err = fmt.Errorf("failed handling forward requests: %w", err)
		return
	}

	return
}

// handleForwardRequests is responsible for handling those
//
func (server *server) handleSSHRequests(
	ctx context.Context,
	conn *ssh.ServerConn,
	reqs <-chan *ssh.Request,
) (err error) {

	var (
		forwarded bool
		logger    = log.WithFields(log.Fields{
			"remote": conn.RemoteAddr().String(),
		})
	)

	// consume the requests sent by the client over the session.
	//
	for r := range reqs {
		logger = logger.WithFields(log.Fields{"type": r.Type})

		switch r.Type {
		case "tcpip-forward":
			if forwarded {
				logger.Info("rejecting request - already forwarded")
				r.Reply(false, nil)
			}

			listener, err := server.listenOnPortForwarded(ctx, conn, r)
			if err != nil {
				err = fmt.Errorf("failed to serve port forward req: %w", err)
				r.Reply(false, nil)
				continue
			}

			defer listener.Close()

			r.Reply(true, ssh.Marshal(tcpipForwardResponse{
				BoundPort: uint32(1234),
			}))

			forwarded = true

		default:
			// aside from keepalives, ignore anything else.
			//
			if !strings.Contains(r.Type, "keepalive") {
				logger.Warn("ignoring")
				r.Reply(false, nil)
				continue
			}

			logger.Debug("keepalive")
			r.Reply(true, nil)
		}
	}

	return
}

func (server *server) listenOnPortForwarded(
	ctx context.Context,
	conn *ssh.ServerConn,
	r *ssh.Request,
) (listener net.Listener, err error) {

	var (
		drain = make(chan struct{})
		wait  = new(sync.WaitGroup)
		req   tcpipForwardRequest
	)

	err = ssh.Unmarshal(r.Payload, &req)
	if err != nil {
		err = fmt.Errorf("failed to parse payload as tcpip forward req: %w", err)
		return
	}

	// listen
	//
	listener, err = net.Listen("tcp", "0.0.0.0:"+strconv.Itoa(int(req.BindPort)))
	if err != nil {
		err = fmt.Errorf("failed to listen on addr: %w", err)
		return
	}

	// start the proxying
	//
	wait.Add(1)
	go server.forwardTCPIP(ctx, drain, wait, conn, listener, req.BindIP, uint16(req.BindPort))

	return
}

// forwardTCPIP takes care of proxying the connections that arive at `listener`
// towards `conn` so that it can make its way to the client.
//
func (server *server) forwardTCPIP(
	ctx context.Context,
	drain <-chan struct{},
	connsWg *sync.WaitGroup,
	conn *ssh.ServerConn,
	listener net.Listener,
	forwardIP string,
	forwardPort uint16,
) (err error) {

	var (
		interrupted = false
		done        = make(chan struct{})
	)

	defer connsWg.Done()
	defer close(done)

	go func() {
		select {
		case <-drain:
			interrupted = true
			listener.Close()
		case <-done:
			log.Debug("done")
		}
	}()

	var localConn net.Conn

	for {
		// take connections from the backlog that we have for that
		// listener.
		//
		localConn, err = listener.Accept()
		if err != nil {
			if !interrupted {
				err = fmt.Errorf("failed to accept connection", err)
				return
			}

			break
		}

		connsWg.Add(1)

		go func() {
			defer connsWg.Done()

			forwardLocalConn(
				ctx,
				localConn,
				conn,
				forwardIP,
				forwardPort,
			)
		}()
	}

	return
}

// forwardLocalConn takes care of proxying bytes from the server to the client,
// and vice-versa.
//
// when a connection comes to a port for which remote forwarding has been
// requested, an SSH channel is opened to forward the port to the other side.
//
//
func forwardLocalConn(
	ctx context.Context,
	localConn net.Conn,
	conn *ssh.ServerConn,
	forwardIP string,
	forwardPort uint16,
) {
	defer localConn.Close()

	host, portStr, err := net.SplitHostPort(localConn.RemoteAddr().String())
	if err != nil {
		panic(fmt.Errorf("failed to split host from port for local conn - %w", err))
	}

	port, err := strconv.ParseUint(portStr, 10, 32)
	if err != nil {
		panic(fmt.Errorf("failed to convert port to uint32"))
	}

	req := forwardTCPIPChannelRequest{
		ForwardIP:   forwardIP,
		ForwardPort: uint32(forwardPort),
		OriginIP:    host,
		OriginPort:  uint32(port),
	}

	channel, reqs, err := conn.OpenChannel("forwarded-tcpip", ssh.Marshal(req))
	if err != nil {
		log.Error("failed to open channel", err)
		return
	}
	defer channel.Close()

	go ssh.DiscardRequests(reqs)

	handleTraffic(ctx, localConn, channel)

	return
}

func handleTraffic(ctx context.Context, a, b io.ReadWriteCloser) {
	const numPipes = 2

	wait := make(chan struct{}, numPipes)
	pipe := func(to io.WriteCloser, from io.ReadCloser) {
		defer to.Close()
		defer from.Close()
		defer func() {
			wait <- struct{}{}
		}()

		io.Copy(to, from)
	}

	go pipe(a, b)
	go pipe(b, a)

	done := 0

dance:
	for {
		select {
		case <-wait:
			done++
			if done == numPipes {
				break dance
			}
		case <-ctx.Done():
			break dance
		}
	}
}

func discardChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
	}
}
