package store

import "github.com/parashmaity/fleare/internal/comm"

type IMemory interface {
	// Length returns the number of items currently stored in memory.
	Length() int32

	// Get retrieves the object associated with the given key.
	// Returns a pointer to the object and an error if the key does not exist or retrieval fails.
	Get(key string) (*comm.Object, error)

	// Set stores a key-value pair in memory.
	// Returns an error if the operation fails.
	Set(key string, value []byte, kind Kind) error

	// SetWithTTL stores a key-value pair in memory with a TTL (Time To Live) in seconds.
	// Returns an error if the operation fails.
	SetWithTTL(key string, value []byte, kind Kind, ttlSeconds int64) error

	// Delete removes the object associated with the given key from memory.
	// Returns an error if the key does not exist or deletion fails.
	Delete(key string) (bool, error)

	// TTL returns the remaining time to live of a key in seconds.
	// Returns -1 if the key exists but has no expiration, -2 if the key does not exist.
	TTL(key string) (int64, error)

	// Expire sets a timeout on key. After the timeout has expired, the key will automatically be deleted.
	// Returns true if the timeout was set, false if key does not exist.
	Expire(key string, ttlSeconds int64) (bool, error)

	// CleanupExpired removes all expired keys from memory.
	// Returns the number of keys that were removed.
	CleanupExpired() int32
}
