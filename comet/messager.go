package comet

import (
	"log"

	nsq "github.com/nsqio/go-nsq"
	"github.com/nsqio/nsq/nsqd"
	"github.com/silentred/beim/lib"
	"github.com/silentred/beim/router"
	"encoding/json"
	"github.com/surgemq/message"
)

type Messager interface {
	router.Router
	// receive msg from MQ
	Receive()
}

type Message struct {
	CometID   string
	UserNames []string
	Message   []byte
}

type CometMessager struct {
	//grpc client to call router
	Router router.Router `inject`
	server *CometServer
	// use to receiing
	nsqCli *nsqd.Consumer
}

func NewCometMessager(server *CometServer) *CometMessager {
	cometTopic := lib.RedisCometTopic(server.ID)
	cfg := nsq.NewConfig()
	consumer, err := nsq.NewConsumer(cometTopic, server.AppName, cfg)
	if err != nil {
		log.Fatal(err)
	}

	cm := &CometMessager{
		server: server,
		nsqCli: consumer,
	}
	
	return cm
}

func (cm *CometMessager) Send(topic string, msg []byte) error {
	return cm.router.Send(topic, msg)
}

func (cm *CometMessager) Receive() {
	cm.nsqCli.AddHandler(cm)
	err := nsqCli.ConnectToNSQD(nsqHost)
	if err != nil {
		log.Fatal(err)
	}

	for {
		select {
		case <-cm.nsqCli.StopChan:
			return
	}
}

func (cm *CometMessager) HandleMessage(msg *nsq.Message) error {
	var m Message
	err := json.Unmarshal(msg.Body, &m)
	if err != nil {
		return err
	}

	for _, username := range m.UserNames {
		if svc, ok := cm.server.clients[username]; ok {
			pubMsg := message.NewPublishMessage()
			_, err := pubMsg.Decode(m.Message)
			if err != nil {
				return err
			}
			svc.writeMsg(pubMsg)
		}
	}
	
	return nil
}
