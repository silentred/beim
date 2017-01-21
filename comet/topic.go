package comet

import (
	"strings"

	"github.com/golang/glog"
	redis "gopkg.in/redis.v5"
)

type TopicProvidor interface {
	Subscribe(clientID, topic string) error
	Unsubscribe(clientID, topic string) error
}

// RedisProvidor
// group_topic_name: [1001, 1002] set for group
// uid_topic_name: No need for individual person, comet knows what topics to send.
type RedisProvidor struct {
	Cli *redis.Client
}

func (r *RedisProvidor) Subscribe(clientID, topic string) {
	if isGroupTopic(topic) {
		err := r.Cli.SAdd(topic, clientID).Err()
		if err != nil {
			glog.Error(err)
		}
	}
}

func (r *RedisProvidor) Unsubscribe(clientID, topic string) {
	if isGroupTopic(topic) {
		err := r.Cli.SRem(topic, clientID).Err()
		if err != nil {
			glog.Error(err)
		}
	}
}

func isGroupTopic(topic string) bool {
	return strings.HasPrefix(topic, "group")
}
