package comet

import (
	"sync"

	"github.com/surgemq/message"
)

const (
	DefaultMaxInflight = 10
)

var memSess = make(map[string]*Session)

type memSessions map[string]*Session

func (ms *memSessions) Create(name string, sess *Session) {
	(*ms)[name] = sess
}
func (ms *memSessions) Update(name string, sess *Session) {
	(*ms)[name] = sess
}
func (ms *memSessions) Delete(name string) {
	delete((*ms), name)
}
func (ms *memSessions) Find(name string) *Session {
	if s, ok := (*ms)[name]; ok {
		return s
	}
	return nil
}

type SessionProvidor interface {
	Create(clientID string) error
	Update(clientID string) error
	Delete(clientID string) error
	Find(clientID string) Session
}

type Session struct {
	// client -> broker message queue
	InStore *MessageStore
	// broker -> client message queue
	OutStore *MessageStore

	// need to be persist
	Persistent bool
	// subscribed topics
	Topics []Topic
}

// MessageStore is a message queue to store the messages to send
type MessageStore struct {
	queue chan *message.PublishMessage
	//contains incoming PacketId, and remove it after handling it
	MsgMap map[uint16]*message.PublishMessage
	//Inflight []uint16
	// max length of Inflight
	maxInflight int
	// for MsgMap
	mutex *sync.RWMutex
}

// NewMessageStore returns a MessageStore
func NewMessageStore(maxInflight int) *MessageStore {
	if maxInflight <= DefaultMaxInflight {
		maxInflight = DefaultMaxInflight
	}

	ms := &MessageStore{
		queue:  make(chan *message.PublishMessage, maxInflight),
		MsgMap: make(map[uint16]*message.PublishMessage),
		mutex:  &sync.RWMutex{},

		maxInflight: maxInflight,
	}

	return ms
}

// Unshift publishMsg to list and wake up waiting goroutine
func (ms *MessageStore) Unshift(msg *message.PublishMessage) {
	// save msg to map
	ms.mutex.Lock()
	ms.MsgMap[msg.PacketId()] = msg
	ms.mutex.Unlock()

	// push it into channel
	// TODO: timeout; get Timer from Pool
	ms.queue <- msg
}

func (ms *MessageStore) Pop() *message.PublishMessage {
	// get msg from channel
	// TODO: timeout
	msg := <-ms.queue

	// remove it from map
	ms.mutex.Lock()
	delete(ms.MsgMap, msg.PacketId())
	ms.mutex.Unlock()

	return msg
}
