//////////////////////////////////////////////////////////////////////
//
// Given is a SessionManager that stores session information in
// memory. The SessionManager itself is working, however, since we
// keep on adding new sessions to the manager our program will
// eventually run out of memory.
//
// Your task is to implement a session cleaner routine that runs
// concurrently in the background and cleans every session that
// hasn't been updated for more than 5 seconds (of course usually
// session times are much longer).
//
// Note that we expect the session to be removed anytime between 5 and
// 7 seconds after the last update. Also, note that you have to be
// very careful in order to prevent race conditions.
//

package main

import (
	"container/list"
	"errors"
	"log"
	"sync"
	"time"
)

type KeySession struct {
	key     string
	session *Session
}

// SessionManager keeps track of all sessions from creation, updating
// to destroying.
type SessionManager struct {
	sync.Mutex
	sessions map[string]*list.Element
	lruKeys  list.List
}

// Session stores the session's data
type Session struct {
	data  map[string]interface{}
	birth time.Time
}

// NewSessionManager creates a new sessionManager
func NewSessionManager() *SessionManager {
	m := &SessionManager{
		sessions: make(map[string]*list.Element),
	}

	go func() {
		for {
			m.cleanSessions()
			time.Sleep(200 * time.Millisecond)
		}
	}()

	return m
}

func (m *SessionManager) cleanSessions() {
	m.Lock()
	defer m.Unlock()
	for cur := m.lruKeys.Back(); cur != nil; {
		age := time.Since(cur.Value.(KeySession).session.birth)
		if age >= 5.0*time.Second {
			m.lruKeys.Remove(cur)
			delete(m.sessions, cur.Value.(KeySession).key)
			cur = cur.Prev()
		} else {
			return
		}
	}
}

// CreateSession creates a new session and returns the sessionID
func (m *SessionManager) CreateSession() (string, error) {
	sessionID, err := MakeSessionID()
	if err != nil {
		return "", err
	}

	m.Lock()
	defer m.Unlock()

	session := Session{
		data:  make(map[string]interface{}),
		birth: time.Now(),
	}
	m.lruKeys.PushFront(KeySession{sessionID, &session})
	m.sessions[sessionID] = m.lruKeys.Front()

	return sessionID, nil
}

// ErrSessionNotFound returned when sessionID not listed in
// SessionManager
var ErrSessionNotFound = errors.New("SessionID does not exists")

// GetSessionData returns data related to session if sessionID is
// found, errors otherwise
func (m *SessionManager) GetSessionData(sessionID string) (map[string]interface{}, error) {
	m.Lock()
	defer m.Unlock()

	session, ok := m.sessions[sessionID]
	if !ok {
		return nil, ErrSessionNotFound
	}
	return session.Value.(KeySession).session.data, nil
}

// UpdateSessionData overwrites the old session data with the new one
func (m *SessionManager) UpdateSessionData(sessionID string, data map[string]interface{}) error {
	m.Lock()
	defer m.Unlock()

	e, ok := m.sessions[sessionID]
	if !ok {
		return ErrSessionNotFound
	}

	// Hint: you should renew expiry of the session here
	m.lruKeys.MoveToFront(e)
	session := e.Value.(KeySession).session
	session.data = data
	session.birth = time.Now()

	return nil
}

func main() {
	// Create new sessionManager and new session
	m := NewSessionManager()
	sID, err := m.CreateSession()
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Created new session with ID", sID)

	// Update session data
	data := make(map[string]interface{})
	data["website"] = "longhoang.de"

	err = m.UpdateSessionData(sID, data)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Update session data, set website to longhoang.de")

	// Retrieve data from manager again
	updatedData, err := m.GetSessionData(sID)
	if err != nil {
		log.Fatal(err)
	}

	log.Println("Get session data:", updatedData)
}
