package ssh

import (
	"context"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"
	"golang.org/x/crypto/ssh"
)

// maxForwards corresponds to the maximum number of port forwarding requests to
// be served for a given SSH connection.
//
const maxForwards = 2

type server struct {
	address string
	config  *ssh.ServerConfig
	port    uint16
}

func NewServer(address, pkey string, port uint16) (s server, err error) {
	config, err := sshServerConfig(pkey)
	if err != nil {
		err = fmt.Errorf("failed generating ssh server config: %w")
		return
	}

	s = server{
		address: address,
		config:  config,
		port:    port,
	}
	return
}

// Server serves an SSH server on a given address.
//
func (server *server) Start(ctx context.Context) (err error) {

	// listen on a given address
	//
	listener, err := net.Listen("tcp", server.address)
	if err != nil {
		err = fmt.Errorf("failed to listen on %s: %w", server.address, err)
		return
	}

	for {
		// wait for someone to go through the tcp connection flow
		//
		c, err := listener.Accept()
		if err != nil {
			if !strings.Contains(err.Error(), "use of closed network connection") {
				log.Error("failed to accept", err)
			}

			continue
		}

		go server.handshake(ctx, c)
	}
}

type ForwardedTCPIP struct {
	BindAddr  string
	BoundPort uint32

	Drain chan<- struct{}

	wg *sync.WaitGroup
}

type ConnState struct {
	ForwardedTCPIPs <-chan ForwardedTCPIP
}

func sshServerConfig(filepath string) (cfg *ssh.ServerConfig, err error) {
	privateBytes, err := ioutil.ReadFile(filepath)
	if err != nil {
		err = fmt.Errorf("failed to load private key from %s: %w", err)
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

// handshake establishes the application-level (SSH) connection.
//
func (server *server) handshake(ctx context.Context, netConn net.Conn) {
	logger := log.WithFields(log.Fields{
		"remote": netConn.RemoteAddr().String(),
	})

	conn, chans, reqs, err := ssh.NewServerConn(netConn, server.config)
	if err != nil {
		logger.Info("handshake failed", err)
		return
	}

	defer conn.Close()

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	forwardedTCPIPs := make(chan ForwardedTCPIP, maxForwards)
	go server.handleGlobalRequests(ctx, conn, reqs, forwardedTCPIPs)

	state := ConnState{
		ForwardedTCPIPs: forwardedTCPIPs,
	}

	chansGroup := new(sync.WaitGroup)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			logger.WithFields(log.Fields{
				"type": newChannel.ChannelType(),
			}).Info("rejecting unknown channel type")

			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			logger.Error("failed to accept channel", err)
			return
		}

		chansGroup.Add(1)
		go server.handleChannel(chansGroup, channel, requests, state)
	}

	chansGroup.Wait()
}

func (server *server) handleChannel(
	chansGroup *sync.WaitGroup,
	channel ssh.Channel,
	requests <-chan *ssh.Request,
	state ConnState,
) {
	return
}

// tcpipForwardRequest is the request sent by the SSH client to request the
// server to start a proxy that is meant to forward requests to a given port.
//
type tcpipForwardRequest struct {

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
type tcpipForwardResponse struct {

	// BoundPort is the port that was created on the server-side to bind to
	// the client's port.
	//
	BoundPort uint32
}

// handleGlobalRequests is responsible for handling those
//
func (server *server) handleGlobalRequests(
	ctx context.Context,
	conn *ssh.ServerConn,
	reqs <-chan *ssh.Request,
	forwardedTCPIPs chan<- ForwardedTCPIP,
) {
	var (
		forwardedThings = 0
		logger          = log.WithFields(log.Fields{
			"remote": conn.RemoteAddr().String(),
		})
	)

	// consume the requests sent by the client over the session.
	//
	for r := range reqs {
		logger = logger.WithFields(log.Fields{"type": r.Type})

		switch r.Type {
		case "tcpip-forward":

			// ensure we don't forward too much
			//
			forwardedThings++
			if forwardedThings > maxForwards {
				logger.Info("rejecting request")
				r.Reply(false, nil)
			}

			// parse the request
			//
			var req tcpipForwardRequest
			err := ssh.Unmarshal(r.Payload, &req)
			if err != nil {
				logger.Error("failed to unmarshal payload as tcpip forward request", err)
				r.Reply(false, nil)
				continue
			}

			// listen on an ephemeral port
			//
			listener, err := net.Listen("tcp", "0.0.0.0:0")
			if err != nil {
				logger.Error("failed to listen", err)
				r.Reply(false, nil)
				continue
			}

			defer listener.Close()

			// get the port that we (server) bound to
			//
			_, port, err := net.SplitHostPort(listener.Addr().String())
			if err != nil {
				panic(fmt.Errorf("failed to split addr generated by net pkg - %w", err))
			}

			// respond back to the client with the port that we've
			// bound to.
			//
			var res tcpipForwardResponse
			_, err = fmt.Sscanf(port, "%d", &res.BoundPort)
			if err != nil {
				panic(fmt.Errorf("failed to retrieve port from string - %w", err))
			}

			logger.Debug("listening")

			forPort := req.BindPort
			if forPort == 0 {
				forPort = res.BoundPort
			}

			drain := make(chan struct{})
			wait := new(sync.WaitGroup)

			wait.Add(1)
			go server.forwardTCPIP(ctx, drain, wait, conn, listener, req.BindIP, forPort)

			// bindAddr := net.JoinHostPort(req.BindIP, fmt.Sprintf("%d", req.BindPort))

		// aside from keepalives, ignore anything else.
		//
		default:
			if strings.Contains(r.Type, "keepalive") {
				logger.Debug("keepalive")
				r.Reply(true, nil)
			} else {
				logger.Warn("ignoring")
				r.Reply(false, nil)
			}
		}
	}

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
	forwardPort uint32,
) {
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

	for {
		// take connections from the backlog that we have for that
		// listener.
		//
		localConn, err := listener.Accept()
		if err != nil {
			if !interrupted {
				log.Error("failed to accept connection", err)
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

type forwardTCPIPChannelRequest struct {
	ForwardIP   string
	ForwardPort uint32
	OriginIP    string
	OriginPort  uint32
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
	forwardPort uint32,
) {
	defer localConn.Close()

	var req forwardTCPIPChannelRequest
	req.ForwardIP = forwardIP
	req.ForwardPort = forwardPort

	host, port, err := net.SplitHostPort(localConn.RemoteAddr().String())
	if err != nil {
		panic(fmt.Errorf("failed to split host from port for local conn - %w", err))
	}

	req.OriginIP = host

	_, err = fmt.Sscanf(port, "%d", &req.OriginPort)
	if err != nil {
		panic(fmt.Errorf("failed to get port from parsed port - %w", err))
	}

	channel, reqs, err := conn.OpenChannel("forwarded-tcpip", ssh.Marshal(req))
	if err != nil {
		log.Error("failed to open channel", err)
		return
	}

	defer channel.Close()
	go ssh.DiscardRequests(reqs)

	return
}
