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
	Set(key string, value []byte) error

	// Delete removes the object associated with the given key from memory.
	// Returns an error if the key does not exist or deletion fails.
	Delete(key string) (bool, error)
}
