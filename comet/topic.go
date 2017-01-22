package comet

import (
	"strings"

	"github.com/golang/glog"
	redis "gopkg.in/redis.v5"
)

type TopicProvidor interface {
	Subscribe(clientID, topic string) error
	Unsubscribe(clientID, topic string) error
	GetGroupMembers(topic string) []string
}

// RedisProvidor
// group_topic_name: [1001, 1002] set for group
// uid_topic_name: No need for individual person, comet knows what topics to send.
type RedisTopic struct {
	Cli *redis.Client
}

func (r *RedisTopic) Subscribe(clientID, topic string) error {
	if isGroupTopic(topic) {
		err := r.Cli.SAdd(topic, clientID).Err()
		if err != nil {
			glog.Error(err)
			return err
		}
	}
	return nil
}

func (r *RedisTopic) Unsubscribe(clientID, topic string) error {
	if isGroupTopic(topic) {
		err := r.Cli.SRem(topic, clientID).Err()
		if err != nil {
			glog.Error(err)
			return err
		}
	}
	return nil
}

func (r *RedisTopic) GetGroupMembers(topic string) []string {
	members, err := r.Cli.SMembers(topic).Result()
	if err != nil {
		glog.Error(err)
		return []string{}
	}
	return members
}

func isGroupTopic(topic string) bool {
	return strings.HasPrefix(topic, "group")
}

type memTopic struct {
	groupUser map[string][]string
}
