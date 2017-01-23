package comet

import (
	redis "gopkg.in/redis.v5"
)

type SessionProvidor interface {
	Create(clientID string, session Session) error
	Update(clientID string, session Session) error
	Delete(clientID string, session Session) error
	Find(clientID string) Session
	AppendOfflineMsg(clientID string, msg []byte) error
	GetOfflineMsg(clientID string, limit int) (bool, error)
	SetOnline(clientID string) error
	SetOffline(clientID string) error
	SetCometID(clientID, cometID string) error
}

type Session struct {
	Username string
	// need to be persist
	Online bool
	// which comet this user is connecting
	CometID    string
	OfflineMsg [][]byte
	// subscribed topics
	TopicNames []string
	TopicQoS   []byte
}

// RedisSession
// Online, CometID, OfflineMsg 频繁更新
// Username, TopicNames, TopicQoS 放在一个 hash 中
type RedisSession struct {
	redisCli *redis.Client
}

// memory session manager
type memSessions struct {
	sessions map[string]Session
}

func newMemSess() *memSessions {
	return &memSessions{
		sessions: new(map[string]Session),
	}
}

func (ms *memSessions) Create(clientID string, sess Session) error {
	ms.sessions[clientID] = sess
	return nil
}
func (ms *memSessions) Update(clientID string, sess Session) error {
	ms.sessions[clientID] = sess
	return nil
}
func (ms *memSessions) Delete(clientID string) error {
	delete(ms.sessions, clientID)
	return nil
}
func (ms *memSessions) Find(clientID string) Session {
	if s, ok := ms.sessions[clientID]; ok {
		return s
	}
	return Session{}
}

func (ms *memSessions) AppendOfflineMsg(clientID string, msg []byte) error {
	s := ms.Find(clientID)
	s.OfflineMsg = append(s.OfflineMsg, msg)
	return ms.Update(clientID, s)
}

func (ms *memSessions) SetOnline(clientID string) error {
	s := ms.Find(clientID)
	s.Online = true
	return ms.Update(clientID, s)
}

func (ms *memSessions) SetOffline(clientID string) error {
	s := ms.Find(clientID)
	s.Online = false
	return ms.Update(clientID, s)
}

func (ms *memSessions) SetCometID(clientID, cometID string) error {
	s := ms.Find(clientID)
	s.CometID = cometID
	return ms.Update(clientID, s)
}
