package wal

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/helper"
	"github.com/parashmaity/fleare/internal/logger"
	"google.golang.org/protobuf/proto"
)

const (
	MaxBufferSize int32  = 10 * 1024 * 1024
	FileName      string = "data-file"
	FileExt       string = "db"
)

type WALShard struct {
	mu          sync.Mutex
	id          string
	file        *os.File
	writer      *bufio.Writer
	batch       [][]byte
	batchSize   int
	flushTicker *time.Ticker
	done        chan struct{}
	writeChan   chan *comm.WALEntry
}

type WALSystem struct {
	shards     map[string]*WALShard
	chash      *helper.ConsistentHash
	writeCount uint64
	entryPool  sync.Pool
}

// NewWALSystem creates a new WAL system
func NewWALSystem(shardCount, batchSize int, walPath string) (*WALSystem, error) {
	err := os.MkdirAll(walPath, 0755) // Changed permissions to 0755
	if err != nil {
		return nil, fmt.Errorf("failed to create WAL directory: %w", err)
	}

	ws := &WALSystem{
		shards: make(map[string]*WALShard),
		entryPool: sync.Pool{
			New: func() interface{} { return new(comm.WALEntry) },
		},
		chash: helper.NewConsistentHash(),
	}

	for i := range shardCount { // Fixed loop range syntax
		id := strconv.Itoa(i)
		shardPath := fmt.Sprintf("%s/%s-%s.%s", walPath, FileName, id, FileExt)
		shard, err := newWALShard(id, batchSize, shardPath)
		if err != nil {
			return nil, err
		}
		ws.shards[id] = shard
		ws.chash.AddNode(id)
	}
	return ws, nil
}

func (ws *WALSystem) Close() error {
	var firstErr error
	for _, shard := range ws.shards {
		if shard != nil {
			if err := shard.Close(); err != nil && firstErr == nil {
				firstErr = err
			}
		}
	}

	return firstErr
}

func (shard *WALShard) Close() error {
	close(shard.done)
	shard.flushTicker.Stop()

	// Ensure all pending writes are flushed
	shard.mu.Lock()
	defer shard.mu.Unlock()

	shard.flushBatch()
	if err := shard.writer.Flush(); err != nil {
		return err
	}
	if err := shard.file.Sync(); err != nil {
		return err
	}
	return shard.file.Close()
}

func newWALShard(shardId string, batchSize int, path string) (*WALShard, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriterSize(file, int(MaxBufferSize))
	shard := &WALShard{
		file:        file,
		id:          shardId,
		writer:      writer,
		batchSize:   batchSize,
		batch:       make([][]byte, 0, batchSize*2), // Pre-allocate batch capacity
		flushTicker: time.NewTicker(30 * time.Second),
		done:        make(chan struct{}),
		writeChan:   make(chan *comm.WALEntry, 10000),
	}

	go shard.processWrites()
	return shard, nil
}

func (ws *WALSystem) GetShard(name string) Shard {
	return ws.shards[name]
}

func (ws *WALSystem) GetShardByKey(key string) Shard {
	node := ws.chash.GetNode(key)
	return ws.shards[node]
}

// Put stores or updates an object
func (ws *WALSystem) Put(key string, obj *comm.Object) error {
	atomic.AddUint64(&ws.writeCount, 1)

	entry := ws.entryPool.Get().(*comm.WALEntry)
	entry.Op = comm.WALEntry_UPDATE
	entry.Key = key
	entry.Object = obj
	entry.Checksum = calculateChecksum(obj)

	// Async write to WAL
	shard := ws.GetShardByKey(key).(*WALShard)
	shard.writeChan <- entry

	return nil
}

// Delete removes an object
func (ws *WALSystem) Delete(key string) error {
	atomic.AddUint64(&ws.writeCount, 1)

	entry := ws.entryPool.Get().(*comm.WALEntry)
	entry.Op = comm.WALEntry_DELETE
	entry.Key = key
	entry.Object = nil
	entry.Checksum = 0

	shard := ws.GetShardByKey(key).(*WALShard)
	shard.writeChan <- entry

	return nil
}

func (shard *WALShard) processWrites() {
	for {
		select {
		case entry := <-shard.writeChan:
			shard.mu.Lock()
			data, err := proto.Marshal(entry)
			if err != nil {
				logger.Error("Protobuf marshal error", err, map[string]any{"entry": entry})
				shard.mu.Unlock()
				continue
			}

			// Write length prefix
			lengthBuf := make([]byte, 4)
			binary.BigEndian.PutUint32(lengthBuf, uint32(len(data)))
			shard.batch = append(shard.batch, lengthBuf)

			// Write actual data
			shard.batch = append(shard.batch, data)

			if len(shard.batch)/2 >= shard.batchSize { // Each entry is 2 slices (length + data)
				shard.flushBatch()
			}
			shard.mu.Unlock()

		case <-shard.flushTicker.C:
			shard.mu.Lock()
			if len(shard.batch) > 0 {
				shard.flushBatch()
			}
			shard.mu.Unlock()

		case <-shard.done:
			shard.mu.Lock()
			shard.flushBatch()
			shard.writer.Flush()
			shard.file.Sync()
			shard.mu.Unlock()
			return
		}
	}
}

func (shard *WALShard) flushBatch() {
	if len(shard.batch) == 0 {
		return
	}

	for _, data := range shard.batch {
		if _, err := shard.writer.Write(data); err != nil {
			logger.Error("WAL write error", err, nil)
		}
	}

	if err := shard.writer.Flush(); err != nil {
		logger.Error("WAL flush error", err, nil)
	}

	// Sync to disk to ensure data is written
	if err := shard.file.Sync(); err != nil {
		logger.Error("WAL sync error", err, nil)
	}

	shard.batch = shard.batch[:0] // Clear batch
}

func calculateChecksum(obj *comm.Object) uint32 {
	if obj == nil {
		return 0
	}
	h := fnv.New32a()
	h.Write(fmt.Appendf(nil, "%d", obj.Kind))
	h.Write(obj.Value)
	h.Write(fmt.Appendf(nil, "%d", obj.Timestamp))
	return h.Sum32()
}

// Stats returns current WAL statistics
func (ws *WALSystem) Stats() uint64 {
	return atomic.LoadUint64(&ws.writeCount)
}

// Recover rebuilds the in-memory map from WAL files
func (ws *WALSystem) Recover(walPath string, callback func(key string, obj *comm.Object, isDeleted bool)) error {
	for i := range len(ws.shards) {
		id := strconv.Itoa(i)
		shardPath := fmt.Sprintf("%s/%s-%s.%s", walPath, FileName, id, FileExt)
		if err := ws.recoverShard(shardPath, callback); err != nil {
			return err
		}
	}
	return nil
}

func (ws *WALSystem) recoverShard(path string, callback func(key string, obj *comm.Object, isDeleted bool)) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // New WAL, nothing to recover
		}
		return err
	}
	defer file.Close()

	reader := bufio.NewReader(file)
	for {
		// Read length prefix
		lengthBuf := make([]byte, 4)
		if _, err := io.ReadFull(reader, lengthBuf); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		length := binary.BigEndian.Uint32(lengthBuf)
		data := make([]byte, length)
		if _, err := io.ReadFull(reader, data); err != nil {
			return err
		}

		var entry comm.WALEntry
		if err := proto.Unmarshal(data, &entry); err != nil {
			return err
		}

		// Verify checksum for data integrity
		if entry.Op == comm.WALEntry_UPDATE && entry.Checksum != calculateChecksum(entry.Object) {
			return fmt.Errorf("checksum mismatch for key %s", entry.Key)
		}

		switch entry.Op {
		case comm.WALEntry_UPDATE:
			callback(entry.Key, entry.Object, false)
		case comm.WALEntry_DELETE:
			callback(entry.Key, entry.Object, true)
		}
	}
	return nil
}
