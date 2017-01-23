package comet

import (
	"errors"
	"fmt"
	"net"
	"time"

	"flag"

	"github.com/golang/glog"
	"github.com/silentred/beim/lib"
	"github.com/surgemq/message"
	redis "gopkg.in/redis.v5"
)

var (
	Injector = lib.NewInjector()

	Store = &lib.Map{}

	errAccept = errors.New("error when acceptting Connection")
	errNoAuth = errors.New("no authrization")

	redisHost string
	redisDB   int
	redisPwd  string

	nsqHost string
)

func init() {
	flag.StringVar(&redisHost, "redis-host", "127.0.0.1:2379", "default localhost:2379")
	flag.IntVar(&redisDB, "redis-db", 1, "default 1")
	flag.StringVar(&redisPwd, "redis-pwd", "", "default empty string")
	flag.StringVar(&nsqHost, "nsq-host", "127.0.0.1:4150", "nsqd TCP interface")
}

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
	ID      string
	AppName string
	// connection count
	ConnCount uint64
	// to validate username/password
	Auth Authenticator
	// send msg to router; receive from MQ
	Msger Messager
	// manage session
	SessManager SessionProvidor
	// topic
	TopicManager TopicProvidor

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

func initServer() {
	// redis client
	redisCli := newRedisClient()
	Store.Set("redis", redisCli)
	Injector.Map(redisCli)

	// auth
	auth := mockAuth{}
	Injector.Map(auth)
	Injector.MapTo(auth, new(Authenticator))

	// session manager
	memSess := newMemSess()
	Injector.Map(memSess)
	Injector.MapTo(memSess, new(SessionProvidor))

	// topic manager
	memTopic := newMemTopic()
	Injector.Map(memTopic)
	Injector.MapTo(memTopic, new(TopicProvidor))

}

func newRedisClient() *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:     redisHost,
		DB:       redisDB,
		Password: redisPwd, // no password set
	})
	if err := client.Ping().Err(); err != nil {
		panic(err)
	}
	return client
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
	username := string(connMsg.Username())
	password := string(connMsg.Password())
	if !server.Auth.Authenticate(username, password) {
		glog.Error(errNoAuth)
		return
	}

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

	// find session

	// set online
	err = server.SessManager.SetOnline(username)
	err = server.SessManager.SetCometID(username, server.ID)
	if err != nil {
		glog.Error(err)
	}

	// send offline msg

	// start service
	svc := newService(conn, DefaultKeepAlive, connMsg, server)
	// save to map
	server.clients[username] = svc
	// for loop service
	svc.start()
}
