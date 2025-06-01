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

	SessionsStore.Store[session.SessionID] = session
	return true, nil
}

func (session *Session) Expire() {
	delete(SessionsStore.Store, session.SessionID)
}

func ExpireById(id string) {
	if SessionsStore.Store[id] != nil {
		SessionsStore.Store[id].Expire()
	}
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
