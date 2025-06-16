package shard

import (
	"fmt"
	"strconv"
	"time"
	"unsafe"

	"github.com/parashmaity/fleare/config"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/helper"
	"github.com/parashmaity/fleare/internal/logger"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/wal"
)

type ShardManager struct {
	shards map[string]*Shard

	chash *helper.ConsistentHash
	wal   wal.WAL
}

func NewShardManager(numShards int) *ShardManager {

	ch := helper.NewConsistentHash()
	w, err := wal.NewWALSystem(numShards, config.GetConfig().Persistence.AfterWriteCount, config.GetConfig().Persistence.Path)
	if err != nil {
		logger.Error("Failed to start persistence service", err, nil)
	}
	shards := make(map[string]*Shard)
	for i := range numShards {

		id := strconv.Itoa(i)
		sh := NewShard(
			strconv.Itoa(i),
			fmt.Sprintf("Shard %d", i),
			fmt.Sprintf("localhost:%d", 9291),
		)

		shards[id] = sh
		ch.AddNode(id)
	}

	return &ShardManager{
		shards: shards,
		chash:  ch,
		wal:    w,
	}
}

func (sm *ShardManager) Wal() wal.WAL {
	return sm.wal
}

func (sm *ShardManager) GetShard(name string) *Shard {
	return sm.shards[name]
}

func (sm *ShardManager) GetShardByKey(key string) *Shard {
	node := sm.chash.GetNode(key)
	return sm.shards[node]
}

func (sm *ShardManager) GetShardCount() int {
	return len(sm.shards)
}

func (sm *ShardManager) GetAllShards() []*Shard {
	shards := make([]*Shard, 0, len(sm.shards))
	for _, shard := range sm.shards {
		shards = append(shards, shard)
	}
	return shards
}

func (sm *ShardManager) GetAllShardMap() map[string]*Shard {
	return sm.shards
}

func (sm *ShardManager) GetAllShardInfo() map[string]any {
	info := make(map[string]any)
	info["shard_count"] = sm.GetShardCount()

	for name, shard := range sm.shards {
		sinfo := make(map[string]any)
		size := unsafe.Sizeof(shard)
		sinfo["key_length"] = shard.M.Length()
		sinfo["total_size"] = fmt.Sprintf("%d bytes", size)
		sinfo["name"] = shard.Name
		sinfo["ID"] = shard.ID
		sinfo["host_address"] = shard.HostAddress
		info[name] = sinfo
	}
	return info
}

func (sm *ShardManager) RecoverMemory(path string) {

	sm.Wal().Recover(path, func(key string, obj *comm.Object, isDeleted bool) {
		shard := sm.GetShardByKey(key)

		if isDeleted {
			shard.M.Delete(key)
		} else {
			if obj == nil {
				return
			}
			if obj.Timestamp > 0 {
				shard.M.SetWithTTL(key, obj.Value, store.Kind(obj.Kind), (obj.Timestamp - time.Now().Unix()))
			} else {
				shard.M.Set(key, obj.Value, store.Kind(obj.Kind))
			}
		}
	})
}
