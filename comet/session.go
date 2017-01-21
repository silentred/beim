package comet

import "github.com/surgemq/message"

var memSess = make(map[string]Session)

type memSessions map[string]Session

func (ms *memSessions) Create(name string, sess Session) {
	(*ms)[name] = sess
}
func (ms *memSessions) Update(name string, sess Session) {
	(*ms)[name] = sess
}
func (ms *memSessions) Delete(name string) {
	delete((*ms), name)
}
func (ms *memSessions) Find(name string) Session {
	if s, ok := (*ms)[name]; ok {
		return s
	}
	return Session{}
}

type SessionProvidor interface {
	Create(clientID string, session Session) error
	Update(clientID string, session Session) error
	Delete(clientID string, session Session) error
	Find(clientID string) Session
	AppendOfflineMsg(clientID string, msg []byte) error
}

type Session struct {
	// need to be persist
	Online     bool
	Persistent bool
	OfflineMsg []message.PublishMessage
	// subscribed topics
	TopicNames []string
	TopicQoS   []byte
}
