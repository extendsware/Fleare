package store

import (
	"fmt"
	"sync"
	"time"

	"github.com/parashmaity/fleare/internal/comm"
)

type Memory struct {
	Mem map[string]*comm.Object
	mu  sync.RWMutex
}

func NewMemory() *Memory {
	return &Memory{
		Mem: make(map[string]*comm.Object),
		mu:  sync.RWMutex{},
	}
}

func (m *Memory) Length() int32 {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return int32(len(m.Mem))
}

func (m *Memory) Get(key string) (*comm.Object, error) {

	m.mu.RLock()
	obj, ok := m.Mem[key]
	m.mu.RUnlock()
	if !ok {
		return nil, nil
	}

	// Check if the object has expired
	if obj.Timestamp > 0 && time.Now().Unix() > obj.Timestamp {
		// Object has expired, remove it and return nil
		delete(m.Mem, key)
		return nil, nil
	}

	return obj, nil
}

func (m *Memory) Set(key string, value []byte, kind Kind) error {

	obj := &comm.Object{Value: value, Kind: uint32(kind)}
	m.mu.Lock()
	m.Mem[key] = obj
	m.mu.Unlock()
	return nil
}

func (m *Memory) SetWithTTL(key string, value []byte, kind Kind, ttlSeconds int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var expirationTime int64 = 0
	if ttlSeconds > 0 {
		expirationTime = time.Now().Unix() + ttlSeconds
	}

	obj := &comm.Object{
		Value:     value,
		Kind:      uint32(kind),
		Timestamp: expirationTime,
	}
	m.Mem[key] = obj
	return nil
}

func (m *Memory) Delete(key string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.Mem[key] != nil {
		delete(m.Mem, key)
		return true, nil
	}
	return false, nil
}

func (m *Memory) TTL(key string) (int64, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	obj, ok := m.Mem[key]
	if !ok {
		return -2, nil // Key does not exist
	}

	if obj.Timestamp == 0 {
		return -1, nil // Key exists but has no expiration
	}

	currentTime := time.Now().Unix()
	if currentTime >= obj.Timestamp {
		// Key has expired, remove it
		delete(m.Mem, key)
		return -2, nil // Key does not exist (expired)
	}
	fmt.Println(obj.Timestamp, currentTime)
	return obj.Timestamp - currentTime, nil
}

func (m *Memory) Expire(key string, ttlSeconds int64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	obj, ok := m.Mem[key]
	if !ok {
		return false, nil // Key does not exist
	}

	// Check if key has already expired
	if obj.Timestamp > 0 && time.Now().Unix() >= obj.Timestamp {
		delete(m.Mem, key)
		return false, nil // Key was expired
	}

	// Set new expiration time
	if ttlSeconds > 0 {
		obj.Timestamp = time.Now().Unix() + ttlSeconds
	} else {
		obj.Timestamp = 0 // Remove expiration
	}

	return true, nil
}

func (m *Memory) CleanupExpired() int32 {
	m.mu.Lock()
	defer m.mu.Unlock()

	currentTime := time.Now().Unix()
	var deletedCount int32 = 0

	for key, obj := range m.Mem {
		if obj.Timestamp > 0 && currentTime >= obj.Timestamp {
			delete(m.Mem, key)
			deletedCount++
		}
	}

	return deletedCount
}
