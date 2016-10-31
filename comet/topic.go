package comet

import (
	"beim/lib"
	"errors"
	"fmt"
)

const (
	stateCHR byte = iota // Regular character
	stateMWC             // Multi-level wildcard
	stateSWC             // Single-level wildcard
	stateSEP             // Topic level separator
	stateSYS             // System level topic ($)
)

const (
	MWC = "#"  // MWC is the multi-level wildcard
	SWC = "+"  // SWC is the single level wildcard
	SEP = "/"  // SEP is the topic level separator
	SYS = "$"  // SYS is the starting character of the system level topics
	_WC = "#+" // Both wildcards
)

var (
	ErrWildCard      = errors.New("Wildcard character '#', '+' must occupy entire topic level")
	ErrWildCardAtEnd = errors.New("Wildcard character '#', '+' must be at the end")
)

type TopicProvider interface {
	Subscribe(topic string, info *SubInfo)
	Unsubscribe(topic, clientID string)
}

// Topic saved in Session
type Topic struct {
	name string
	qos  byte
}

// SubInfo saved in TopicTree
// router 找到所有订阅者后，直接发送到 queue, comet 订阅 clientID topic, 就能接收到msg;
// 目前开来，router 不需要 qos 信息
type SubInfo struct {
	ClientID string
	QoS      byte
}

// TopicTree holds all subscription info
type TopicTree struct {
	root *topicNode
}

func (tt *TopicTree) Subscribe(topic string, info *SubInfo) {

}

func (tt *TopicTree) Unsubscribe(topic, clientID string) {

}

func (tt *TopicTree) GetSubscribers(topic string, result []*SubInfo) {

}

type topicNode struct {
	Name     string
	Infos    []*SubInfo
	children []*topicNode
}

func newNode(name string) *topicNode {
	return &topicNode{}
}

func (tn *topicNode) findOrAppendNode(node *topicNode) {
	// need mutex ??
	child := tn.findChildByName(node.Name)

	if child == nil {
		tn.children = append(tn.children, node)
	}
}

func (tn *topicNode) findChildByName(name string) *topicNode {
	n := len(tn.children)
	for i := 0; i < n; i++ {
		if tn.children[i].Name == name {
			return tn.children[i]
		}
	}

	return nil
}

func (tn *topicNode) findOrCreateChildByName(name string) *topicNode {
	node := tn.findChildByName(name)
	if node == nil {
		node = newNode(name)
	}

	tn.children = append(tn.children, node)

	return node
}

func (tn *topicNode) appendSubInfo(info *SubInfo) {
	// need mutex ??
	var found bool
	n := len(tn.Infos)
	for i := 0; i < n; i++ {
		if tn.Infos[i].ClientID == info.ClientID {
			found = true
			break
		}
	}

	if !found {
		tn.Infos = append(tn.Infos, info)
	}
}

func parseTopic(topic string) (res []string, err error) {
	n := len(topic)
	resBytes := make([][]byte, 1)
	topicBytes := []byte(topic)
	var lastSepIndex int

	for i, c := range topicBytes {
		if c == '/' {
			resBytes = append(resBytes, topicBytes[lastSepIndex:i])
			lastSepIndex = i + 1
		}
		if i == n-1 {
			resBytes = append(resBytes, topicBytes[lastSepIndex:])
		}
	}

	n = len(resBytes)
	for i, nodeName := range resBytes {
		// validate nodeName
		if len(nodeName) > 1 && bytesContains(nodeName, []byte{'+', '#'}) {
			err = ErrWildCard
			return
		}

		if i < n-1 && bytesContains(nodeName, []byte{'+', '#'}) {
			err = ErrWildCardAtEnd
			return
		}

		res = append(res, lib.ByteString(nodeName))
	}

	return
}

// Returns topic level, remaining topic levels and any errors
func nextTopicLevel(topic []byte) ([]byte, []byte, error) {
	s := stateCHR

	for i, c := range topic {
		switch c {
		case '/':
			if s == stateMWC {
				return nil, nil, fmt.Errorf("memtopics/nextTopicLevel: Multi-level wildcard found in topic and it's not at the last level")
			}
			if i == 0 {
				return []byte(""), topic[i+1:], nil
			}
			return topic[:i], topic[i+1:], nil
		case '#':
			if i != 0 {
				return nil, nil, fmt.Errorf("memtopics/nextTopicLevel: Wildcard character '#' must occupy entire topic level")
			}
			s = stateMWC
		case '+':
			if i != 0 {
				return nil, nil, fmt.Errorf("memtopics/nextTopicLevel: Wildcard character '+' must occupy entire topic level")
			}
			s = stateSWC
		case '$':
			s = stateSYS
		default:
			if s == stateMWC || s == stateSWC {
				return nil, nil, fmt.Errorf("memtopics/nextTopicLevel: Wildcard characters '#' and '+' must occupy entire topic level")
			}
			s = stateCHR
		}
	}

	// If we got here that means we didn't hit the separator along the way, so the
	// topic is either empty, or does not contain a separator. Either way, we return
	// the full topic
	return topic, nil, nil
}

func bytesContains(stack []byte, needle []byte) bool {
	for _, s := range stack {
		for _, n := range needle {
			if s == n {
				return true
			}
		}
	}
	return false
}
