package comet

import (
	"errors"
	"fmt"
	"net"
	"time"

	"github.com/golang/glog"
	"github.com/surgemq/message"
)

var (
	errAccept = errors.New("error when acceptting Connection")
	errNoAuth = errors.New("no authrization")
)

const (
	DefaultKeepAlive = 120
)

type Comet interface {
	Serve(net.Listener) error
	HandleConnection(conn net.Conn)
	// dispatch msg to client connection after receiving from MQ
	DispatchMsg()
}

// CometServer of IM
type CometServer struct {
	Config *CometConfig
	// connection count
	ConnCount uint64
	// to validate username/password
	Auth Authenticator
	// send msg to router; receive from
	messager Messager
	// stop signal chan
	stop chan struct{}
	// map of clientID and Service running
	clients map[string]*service
}

func NewServer(config *CometConfig) *CometServer {
	return &CometServer{
		Config:  config,
		Auth:    new(mockAuth),
		stop:    make(chan struct{}, 1),
		clients: make(map[string]*service),
	}
}

func (server *CometServer) ListenAndServe(addr string) (err error) {
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

func (server *CometServer) Serve(l net.Listener) error {
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

		go server.HandleConnection(conn)
	}

	return errAccept
}

func (server *CometServer) HandleConnection(conn net.Conn) {
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
	srv := newService(conn, DefaultKeepAlive, connMsg, server)
	// save to map
	server.clients[string(connMsg.Username())] = srv
	// for loop service
	srv.start()
}
