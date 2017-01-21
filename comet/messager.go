package comet

import (
	"github.com/nsqio/nsq/nsqd"
)

type Messager interface {
	// send msg to router
	KeepSending()
	// receive msg from MQ
	KeepReceiving()
}

type Message struct {
	Topic  string
	ToUser string
	Body   []byte
}

type CometMessager struct {
	server *CometServer
	nsqCli nsqd.Consumer
}
