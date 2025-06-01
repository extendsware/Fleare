package wal

import "github.com/parashmaity/fleare/internal/comm"

// WAL defines the public interface for a Write-Ahead Log system.
type WAL interface {
	Put(key string, obj *comm.Object) error
	Delete(key string) error
	GetShard(name string) Shard
	GetShardByKey(key string) Shard
	Stats() (writeCount uint64)
	Close() error
	Recover(walPath string, callback func(key string, obj *comm.Object, isDeleted bool)) error
}

// Shard defines the minimal behavior expected from a WAL shard.
type Shard interface {
	Close() error
}
