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

func newMemTopic() *memTopic {
	return &memTopic{
		groupUser: new(map[string][]string]),
	}
}

func (mt *memTopic) Subscribe(clientID, topic string) error {
	if unames, ok := mt.groupUser[topic]; ok {
		mt.groupUser[topic] = append(unames, clientID)
	} else {
		mt.groupUser[topic] = []string{clientID}
	}
}

func (mt *memTopic) Unsubscribe(clientID, topic string) error {
	if unames, ok := mt.groupUser[topic]; ok {
		var delIndex int
		for i, val := range unames {
			if val == clientID {
				delIndex = i
				break
			}
		}

		mt.groupUser[topic] = delIndexString(delIndex, mt.groupUser[topic])
	}
}

func (mt *memTopic) GetGroupMembers(topic string) []string {
	if unames, ok := mt.groupUser[topic]; ok {
		return unames
	}
	return []string{}
}

func delIndexString(i int, s []string) []string {
	return s[:i+copy(s[i:], s[i+1:])]
}
