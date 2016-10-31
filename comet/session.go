package comet

import (
	"container/list"
	"sync"

	"github.com/surgemq/message"
)

const (
	DefaultMaxInflight = 10
)

type SessionProvidor interface {
	Create(clientID string) error
	Update(clientID string) error
	Delete(clientID string) error
	Find(clientID string) Session
}

type Session struct {
	// client -> broker message queue
	InQueue *MessageStore
	// broker -> client message queue
	OutQueue *MessageStore

	// need to be persist
	Persistent bool
	// subscribed topics
	Topics []Topic
}

// MessageStore is a message queue to store the messages to send
type MessageStore struct {
	List     *list.List
	Cond     *sync.Cond
	MsgMap   map[uint16]*message.PublishMessage
	Inflight []uint16
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

	l := &sync.Mutex{}
	ms := &MessageStore{
		MsgMap:      make(map[uint16]*message.PublishMessage),
		List:        list.New(),
		Cond:        sync.NewCond(l),
		Inflight:    make([]uint16, maxInflight),
		mutex:       &sync.RWMutex{},
		maxInflight: maxInflight,
	}

	return ms
}

// Wait imcoming msg, then send to Router
func (ms *MessageStore) Wait() {
	ms.Cond.L.Lock()

	for ms.List.Len() == 0 {
		ms.Cond.Wait()
	}

	// can we put this(handling msg) outside in Service???
	// yes. add one function as service.HandleIncomingMsg func (ms *MessageStore , msg *PublishMessage)
	// and pass it to MessageStore
	msg := ms.Pop()

	// TODO: do send to Router

	if msg != nil {
		switch msg.QoS() {
		case message.QosAtLeastOnce:
			// ms.RemoveFromInflight(msg.PacketId())
			fallthrough
		case message.QosExactlyOnce:
			ms.RemoveFromInflight(msg.PacketId())
			// when receiving pubrel from client, service will search pkgID from this store.
			// if failed to find, then send pubcom back.
			// case AtMostOnce:
			// it has sent msg to Router directly in service
		}
	}

	ms.Cond.L.Unlock()
}

// Unshift publishMsg to list and wake up waiting goroutine
func (ms *MessageStore) Unshift(msg *message.PublishMessage) {
	// insert to map
	ms.mutex.Lock()

	// limit the maxLen of List
	if ms.List.Len() > 32 {
		ms.Cond.Wait()
	}

	ms.MsgMap[msg.PacketId()] = msg

	// unshift to list
	ms.List.PushFront(&list.Element{Value: msg.PacketId()})

	// wake up one goroutine
	ms.Cond.Signal()

	ms.mutex.Unlock()
}

func (ms *MessageStore) Pop() *message.PublishMessage {
	// pop from list
	ele := ms.List.Remove(ms.List.Back())
	pkgID, ok := ele.(*list.Element).Value.(uint16)

	if ok {
		if msg, ok := ms.MsgMap[pkgID]; ok {
			// delete from map
			if msg.QoS() == 0 {
				ms.mutex.Lock()
				delete(ms.MsgMap, msg.PacketId())
				ms.mutex.Unlock()
			} else {
				// move msg to Inflight
				ms.moveToInflight(pkgID)
			}
		}
	}

	return nil
}

func (ms *MessageStore) moveToInflight(pkgID uint16) {
	ms.mutex.Lock()

	n := len(ms.Inflight)
	ms.Inflight = append(ms.Inflight, pkgID)

	if n >= ms.maxInflight {
		ms.Inflight = ms.Inflight[ms.maxInflight-n:]
	}

	ms.mutex.Unlock()
}

func (ms *MessageStore) RemoveFromInflight(pkgID uint16) bool {
	ms.mutex.Lock()

	var index int
	var found bool
	for i, val := range ms.Inflight {
		if val == pkgID {
			index = i
			found = true
			break
		}
	}

	// delete index
	//ms.Inflight = append(ms.Inflight[:index], ms.Inflight[index+1:]...)
	if found {
		ms.Inflight = ms.Inflight[:index+copy(ms.Inflight[index:], ms.Inflight[index+1:])]
	}

	ms.mutex.Unlock()

	return found
}
