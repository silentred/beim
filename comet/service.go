package comet

import (
	"fmt"
	"io"
	"sync/atomic"
	"time"

	"github.com/golang/glog"
	"github.com/surgemq/message"
)

var (
	serviceID uint64
)

type service struct {
	// use username as ID
	ID string
	// contains username, password,
	ConnMsg *message.ConnectMessage
	conn    io.Closer
	quit    chan struct{}
	timer   *time.Timer
	server  *CometServer

	keepAlive time.Duration
}

func newService(conn io.Closer, keepAlive int, connMsg *message.ConnectMessage, server *CometServer) *service {
	return &service{
		ID:      string(connMsg.Username()),
		ConnMsg: connMsg,
		conn:    conn,
		quit:    make(chan struct{}, 1),
		server:  server,

		keepAlive: time.Duration(keepAlive) * time.Second,
	}
}

func (service *service) start() {
	// new a timer
	service.timer = time.NewTimer(service.keepAlive)
	// add count
	atomic.AddUint64(&service.server.ConnCount, 1)

	// for loop to get Message
	for {
		select {
		case <-service.quit:
			return
		case <-service.timer.C:
			service.stop()
			continue
		default:
		}

		msg, err := service.getMessage()
		if err != nil {
			glog.Error(err)
			return
		}
		service.handleMsg(msg)

		service.timer.Reset(service.keepAlive)
	}

	// receiving msg From Queue

}

func (service *service) stop() {
	// break for loop
	service.quit <- struct{}{}
	atomic.AddUint64(&service.server.ConnCount, ^uint64(0))

	// recycle resource
	err := service.conn.Close()
	if err != nil {
		glog.Error(err)
	}
}

func (service *service) getMessage() (msg message.Message, err error) {
	buf, err := getMessageBuffer(service.conn)
	if err != nil {
		return
	}

	mType := message.MessageType(buf[0] >> 4)
	msg, err = mType.New()
	if err != nil {
		return
	}

	_, err = msg.Decode(buf)
	if err != nil {
		return
	}

	return
}

func (service *service) handleMsg(msg message.Message) {
	switch msg := msg.(type) {
	case *message.SubscribeMessage:
		service.handleSubscribeMsg(msg)
	case *message.UnsubscribeMessage:
		service.handleUnsubscribeMsg(msg)
	case *message.PublishMessage:
		service.handlePublishMsg(msg)
	case *message.PingreqMessage:
		service.handlePingReq(msg)
	case *message.PubrecMessage:
		service.handlePubrecMsg(msg)
	case *message.PubrelMessage:
		service.handlePubrelMsg(msg)
	case *message.PubcompMessage:
		service.handlePubcomMsg(msg)

	case *message.DisconnectMessage:
		service.stop()
	default:
		fmt.Println(msg)
	}
}

func (service *service) waitQueue() {
	//
}

func (service *service) sendMsgToRouter(b []byte) error {

	return nil
}

func (service *service) handlePingReq(msg *message.PingreqMessage) {
	resp := message.NewPingrespMessage()
	service.writeMsg(resp)
}

func (service *service) handlePublishMsg(msg *message.PublishMessage) {
	switch msg.QoS() {
	case message.QosAtLeastOnce:
		resp := message.NewPubackMessage()
		resp.SetPacketId(msg.PacketId())
		fmt.Println("PacketId ", resp.PacketId())
		service.writeMsg(resp)

	case message.QosExactlyOnce:
		resp := message.NewPubrecMessage()
		resp.SetPacketId(msg.PacketId())
		service.writeMsg(resp)

	case message.QosAtMostOnce:
		// send to Router directly
	}

	// for simplecity, just send to Router
	// just send to destination for test; TODO stick with MQTT
	infos := memTopicTree.GetSubscribers(string(msg.Topic()))
	for _, info := range infos {
		srv := service.server.clientServices[info.ClientID]
		srv.writeMsg(msg)
	}

}

func (service *service) handleSubscribeMsg(msg *message.SubscribeMessage) {
	resp := message.NewSubackMessage()
	resp.SetPacketId(msg.PacketId())

	// topics
	topics := msg.Topics()
	qoss := msg.Qos()

	for i, val := range topics {
		fmt.Printf("topic %s, qos %d \n", string(val), qoss[i])
		memTopicTree.Subscribe(string(val), &SubInfo{string(service.ConnMsg.Username()), qoss[i]})
	}

	// if sub success
	resp.AddReturnCodes(msg.Qos())
	service.writeMsg(resp)
}

func (service *service) handleUnsubscribeMsg(msg *message.UnsubscribeMessage) {
	resp := message.NewUnsubackMessage()
	resp.SetPacketId(msg.PacketId())

	// unsub topics
	topics := msg.Topics()
	for _, val := range topics {
		fmt.Printf("topic %s \n", string(val))
		memTopicTree.Unsubscribe(string(val), string(service.ConnMsg.Username()))
	}

	service.writeMsg(resp)
}

func (service *service) handlePubrecMsg(msg *message.PubrecMessage) {
	resp := message.NewPubrelMessage()
	resp.SetPacketId(msg.PacketId())

	// release message
	service.writeMsg(resp)
}

func (service *service) handlePubrelMsg(msg *message.PubrelMessage) {
	// try to find the msg by packetID. if not found, it means msg has been sent

	resp := message.NewPubcompMessage()
	resp.SetPacketId(msg.PacketId())

	service.writeMsg(resp)
}

func (service *service) handlePubcomMsg(msg *message.PubcompMessage) {
	// delete message from inflight queue

}

func (service *service) writeMsg(msg message.Message) {
	err := writeMessage(service.conn, msg)
	if err != nil {
		glog.Error(err)
	}
}
