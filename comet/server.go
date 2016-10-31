package comet

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/surge/glog"
	"github.com/surgemq/message"
)

var (
	errAccept = errors.New("error when acceptting Connection")
	errNoAuth = errors.New("no authrization")
)

const (
	DefaultKeepAlive = 120
)

// Server of IM
type Server struct {
	Config *ServerConfig
	// to validate username/password
	Auth Authenticator

	// stop signal chan
	stop chan struct{}
	// map of clientID and Service running
	clientServices map[string]*service
}

func NewServer(config *ServerConfig) *Server {
	return &Server{
		Config: config,
		Auth:   new(mockAuth),
		stop:   make(chan struct{}, 1),
	}
}

func (server *Server) ListenAndServe(addr string) (err error) {
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}

	err = server.Serve(l)
	if err != nil {
		return err
	}

	return l.Close()
}

func (server *Server) Serve(l net.Listener) error {
	var tempDelay time.Duration

	for {
		// before Accept(), check if it should stop Server
		select {
		case <-server.stop:
			break
		default:
		}

		conn, err := l.Accept()
		if err != nil {
			if errNet, ok := err.(net.Error); ok && errNet.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				glog.Errorf("server.Serve: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}

			// if it is not netError and not Temporary, stop the server
			glog.Error(err)
			break
		}

		go server.handleConnection(conn)
	}

	return errAccept
}

func (server *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.SetReadDeadline(time.Now().Add(time.Second * 3))

	// handle connect request
	connMsg, err := getConnectMessage(conn)
	if err != nil {
		// may be timeout error
		glog.Error(err)
		return
	}
	fmt.Println("CONN msg", connMsg.String())

	// after receiving the CONN msg, set read deadline to forever
	conn.SetReadDeadline(time.Time{})

	// do auth
	if !server.Auth.Authenticate(string(connMsg.Username()), string(connMsg.Password())) {
		glog.Error(errNoAuth)
		return
	}

	// find session

	// return connack msg
	connAck := message.NewConnackMessage()
	connAck.SetPacketId(connMsg.PacketId())
	connAck.SetSessionPresent(false)
	connAck.SetReturnCode(message.ConnectionAccepted)
	err = writeMessage(conn, connAck)
	if err != nil {
		glog.Error(err)
		return
	}

	// subscribe topics for client
	// send offline

	// start service
	srv := newService(conn, DefaultKeepAlive)
	// for loop service
	srv.start()

}
