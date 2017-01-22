package comet

import "github.com/nsqio/nsq/nsqd"

type Messager interface {
	// send msg to router
	KeepSending()
	// receive msg from MQ
	KeepReceiving()
}

type MsgPublisher interface {
	Publish(SingleMessage) error
	PublishGroup(GroupMessage) error
}

type MsgConsumer interface {
	Handle(SingleMessage) error
	HandleGroup(GroupMessage) error
}

type SingleMessage struct {
	Topic    string
	UserName string
	Message  []byte
}

type GroupMessage struct {
	CometID   string
	UserNames []string
	Message   []byte
}

type CometMessager struct {
	server *CometServer
	nsqCli nsqd.Consumer
	//grpc client to call router
}
