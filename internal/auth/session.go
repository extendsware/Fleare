package auth

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/parashmaity/fleare/internal/utils"
)

// EmptyStr represents an empty string for comparison
const EmptyStr = ""

// Session status constants
const (
	SessionStatusPending = iota
	SessionStatusActive
	SessionStatusExpired
)

type Sessions struct {
	Store  map[string]*Session
	stLock *sync.RWMutex
}

var SessionsStore *Sessions

func init() {
	SessionsStore = NewSessionsStore()
}

func NewSessionsStore() (sessions *Sessions) {
	sessions = &Sessions{
		Store:  make(map[string]*Session),
		stLock: &sync.RWMutex{},
	}
	return
}

type Session struct {
	SessionID      string    `json:"session_id"`
	User           *User     `json:"user"`
	Status         int       `json:"status"`
	CreatedAt      time.Time `json:"created_at"`
	LastAccessedAt time.Time `json:"last_accessed_at"`
}

func NewSession(user *User, clientID string) (session *Session) {
	session = &Session{
		SessionID:      clientID,
		CreatedAt:      utils.GetCurrentTime(),
		LastAccessedAt: utils.GetCurrentTime(),
		User:           user,
		Status:         SessionStatusPending,
	}
	return
}

func (session *Session) Activate() (bool, error) {
	session.Status = SessionStatusActive
	session.CreatedAt = utils.GetCurrentTime().UTC()
	session.LastAccessedAt = utils.GetCurrentTime().UTC()

	SessionsStore.stLock.Lock()
	SessionsStore.Store[session.SessionID] = session
	SessionsStore.stLock.Unlock()
	return true, nil
}

// expireInternal deletes session without locking (internal use)
func expireInternal(sessionID string) {
	delete(SessionsStore.Store, sessionID)
}

func (session *Session) Expire() {
	SessionsStore.stLock.Lock()
	expireInternal(session.SessionID)
	SessionsStore.stLock.Unlock()
}

func ExpireById(id string) {
	SessionsStore.stLock.Lock()
	defer SessionsStore.stLock.Unlock()

	if _, exists := SessionsStore.Store[id]; exists {
		expireInternal(id)
	}
}

// GetById safely retrieves a session by ID
func GetById(id string) *Session {
	SessionsStore.stLock.RLock()
	defer SessionsStore.stLock.RUnlock()

	if session, exists := SessionsStore.Store[id]; exists {
		return session
	}
	return nil
}

// convert to string

func (session *Session) String() (string, error) {
	// Create a copy of the user without password
	userCopy := *session.User
	userCopy.Password = "*******"

	// Create a session copy with the modified user
	sessionCopy := *session
	sessionCopy.User = &userCopy

	jsonData, err := json.Marshal(sessionCopy)
	if err != nil {
		return "", err
	}
	return string(jsonData), nil
}
