package shard

import (
	"sync"

	"github.com/parashmaity/fleare/internal/store"
)

type Shard struct {
	mu          sync.RWMutex
	M           store.IMemory
	Name        string
	ID          string
	HostAddress string
}

func NewShard(id string, name string, hostAddress string) *Shard {
	return &Shard{
		M:           store.NewMemory(),
		ID:          id,
		Name:        name,
		HostAddress: hostAddress,
	}
}
