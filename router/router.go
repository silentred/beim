package router

import (
	"fmt"

	"strings"

	gonsq "github.com/nsqio/go-nsq"
	"github.com/silentred/beim/comet"
	"github.com/silentred/beim/lib"
	"github.com/surge/glog"
)

type Router interface {
	Send(topic string, msg []byte) error
}

type SimpleRouter struct {
	TopicManager comet.TopicProvidor   `inject`
	SessManager  comet.SessionProvidor `inject`

	NsqCli *gonsq.Producer `inject`
}

func NewSimpleRouter(topicMgr comet.TopicProvidor, sessMgr comet.SessionProvidor, nsqCli *gonsq.Producer) *SimpleRouter {
	return &SimpleRouter{
		TopicManager: topicMgr,
		SessManager:  sessMgr,
		NsqCli:       nsqCli,
	}
}

func (r *SimpleRouter) RouteSingleMsg(topic string, byteMsg []byte) error {
	// get receiver username from topic, pattern user/{uid}
	if len(topic) > 6 {
		uid := topic[5:]
		username := fmt.Sprintf("app-%s", uid)
		sess := r.SessManager.Find(username)

		if sess.Online {
			singleMsg := comet.Message{
				CometID:   sess.CometID,
				UserNames: []string{username},
				Message:   byteMsg,
			}
			return r.Publish(singleMsg)
		} else {
			return r.SessManager.AppendOfflineMsg(username, byteMsg)
		}
	}

	return fmt.Errorf("invalid topic for single msg: %s", topic)
}

func (r *SimpleRouter) RouteGroupMsg(topic string, byteMsg []byte) error {
	var messages map[string]comet.Message

	// find uid list
	usernames := r.TopicManager.GetGroupMembers(topic)
	for _, username := range usernames {
		sess := r.SessManager.Find(username)
		if sess.Online {
			if gmsg, ok := messages[sess.CometID]; ok {
				gmsg.UserNames = append(gmsg.UserNames, username)
			} else {
				gmsg = comet.Message{
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
		err := r.Publish(val)
		if err != nil {
			glog.Error(err)
		}
	}

	return nil
}

func (r *SimpleRouter) Publish(msg comet.Message) error {
	cometTopic := lib.RedisCometTopic(msg.CometID)
	return r.NsqCli.Publish(cometTopic, msg.Message)
}

func (r *SimpleRouter) Send(topic string, msg []byte) error {
	if strings.HasPrefix(topic, "group") {
		return r.RouteGroupMsg(topic, msg)
	} else {
		return r.RouteSingleMsg(topic, msg)
	}
}
