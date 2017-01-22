package router

import (
	"fmt"

	"github.com/silentred/beim/comet"
	"github.com/surge/glog"
	"github.com/surgemq/message"
)

type Router interface {
	RouteGroupMsg(*message.PublishMessage) error
}

type SimpleRouter struct {
	TopicManager comet.TopicProvidor
	SessManager  comet.SessionProvidor
	Publisher    comet.MsgPublisher
}

func NewSimpleRouter(topicMgr comet.TopicProvidor, sessMgr comet.SessionProvidor, publisher comet.MsgPublisher) *SimpleRouter {
	return &SimpleRouter{
		TopicManager: topicMgr,
		SessManager:  sessMgr,
		Publisher:    publisher,
	}
}

func (r *SimpleRouter) RouteSingleMsg(msg *message.PublishMessage) error {
	var byteMsg []byte
	msg.Encode(byteMsg)

	topic := string(msg.Topic())
	// get receiver username from topic, pattern user/{uid}
	if len(topic) > 6 {
		uid := topic[5:]
		username := fmt.Sprintf("app-%s", uid)
		sess := r.SessManager.Find(username)

		if sess.Online {
			singleMsg := comet.SingleMessage{
				Topic:    topic,
				UserName: username,
				Message:  byteMsg,
			}
			return r.Publisher.Publish(singleMsg)
		} else {
			return r.SessManager.AppendOfflineMsg(username, byteMsg)
		}
	}

	return fmt.Errorf("invalid topic for single msg: %s", topic)
}

func (r *SimpleRouter) RouteGroupMsg(msg *message.PublishMessage) error {
	var messages map[string]comet.GroupMessage
	topic := string(msg.Topic())
	// find uid list
	usernames := r.TopicManager.GetGroupMembers(topic)
	for _, username := range usernames {
		var byteMsg []byte
		msg.Encode(byteMsg)
		sess := r.SessManager.Find(username)
		if sess.Online {
			if gmsg, ok := messages[sess.CometID]; ok {
				gmsg.UserNames = append(gmsg.UserNames, username)
			} else {
				gmsg = comet.GroupMessage{
					CometID:   sess.CometID,
					Message:   byteMsg,
					UserNames: []string{username},
				}
				messages[sess.CometID] = gmsg
			}
		} else {
			// append offline log
			return r.SessManager.AppendOfflineMsg(username, byteMsg)
		}
	}

	for _, val := range messages {
		err := r.Publisher.PublishGroup(val)
		if err != nil {
			glog.Error(err)
		}
	}

	return nil
}
