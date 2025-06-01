package main

import (
	"bufio"
	"encoding/gob"
	"fmt"
	"hash/fnv"
	"io"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

// LogEntry represents a single operation in the WAL
type LogEntry struct {
	Op        byte // 'U' for update, 'D' for delete
	Key       string
	Value     []byte
	Timestamp int64
	Checksum  uint32 // For data integrity verification
}

// WALShard represents a single shard of the WAL system
type WALShard struct {
	mu          sync.Mutex
	file        *os.File
	writer      *bufio.Writer
	encoder     *gob.Encoder
	batch       []LogEntry
	batchSize   int
	flushTicker *time.Ticker
	done        chan struct{}
	writeChan   chan LogEntry
}

// WALSystem represents the complete WAL system with sharding
type WALSystem struct {
	shards     []*WALShard
	memMap     *sync.Map
	shardCount int
	writeCount uint64 // For monitoring
	flushCount uint64
	entryPool  sync.Pool
}

// NewWALSystem creates a new WAL system
func NewWALSystem(shardCount, batchSize int, walPath string) (*WALSystem, error) {
	ws := &WALSystem{
		shards:     make([]*WALShard, shardCount),
		memMap:     new(sync.Map),
		shardCount: shardCount,
		entryPool: sync.Pool{
			New: func() interface{} { return new(LogEntry) },
		},
	}

	for i := 0; i < shardCount; i++ {
		shardPath := fmt.Sprintf("%s.%d", walPath, i)
		shard, err := newWALShard(batchSize, shardPath)
		if err != nil {
			return nil, err
		}
		ws.shards[i] = shard
	}

	return ws, nil
}

func newWALShard(batchSize int, path string) (*WALShard, error) {
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	writer := bufio.NewWriterSize(file, 64*1024) // 64KB buffer
	shard := &WALShard{
		file:        file,
		writer:      writer,
		encoder:     gob.NewEncoder(writer),
		batchSize:   batchSize,
		flushTicker: time.NewTicker(10 * time.Millisecond),
		done:        make(chan struct{}),
		writeChan:   make(chan LogEntry, 10000), // Buffered channel
	}

	go shard.processWrites()
	return shard, nil
}

func (ws *WALSystem) getShard(key string) *WALShard {
	h := fnv.New32a()
	h.Write([]byte(key))
	return ws.shards[h.Sum32()%uint32(ws.shardCount)]
}

// Put stores a key-value pair
func (ws *WALSystem) Put(key string, value []byte) error {
	atomic.AddUint64(&ws.writeCount, 1)

	// Get entry from pool to reduce allocations
	entry := ws.entryPool.Get().(*LogEntry)
	entry.Op = 'U'
	entry.Key = key
	entry.Value = value
	entry.Timestamp = time.Now().UnixNano()
	entry.Checksum = calculateChecksum(value)

	// Store in memory immediately
	ws.memMap.Store(key, value)

	// Async write to WAL
	shard := ws.getShard(key)
	shard.writeChan <- *entry

	// Return entry to pool
	ws.entryPool.Put(entry)
	return nil
}

// Delete removes a key
func (ws *WALSystem) Delete(key string) error {
	atomic.AddUint64(&ws.writeCount, 1)

	entry := ws.entryPool.Get().(*LogEntry)
	entry.Op = 'D'
	entry.Key = key
	entry.Value = nil
	entry.Timestamp = time.Now().UnixNano()
	entry.Checksum = 0

	ws.memMap.Delete(key)

	shard := ws.getShard(key)
	shard.writeChan <- *entry

	ws.entryPool.Put(entry)
	return nil
}

// Get retrieves a value
func (ws *WALSystem) Get(key string) ([]byte, bool) {
	val, ok := ws.memMap.Load(key)
	if !ok {
		return nil, false
	}
	return val.([]byte), true
}

func (ws *WALShard) processWrites() {
	for {
		select {
		case entry := <-ws.writeChan:
			ws.mu.Lock()
			ws.batch = append(ws.batch, entry)
			if len(ws.batch) >= ws.batchSize {
				ws.flushBatch()
			}
			ws.mu.Unlock()

		case <-ws.flushTicker.C:
			ws.mu.Lock()
			if len(ws.batch) > 0 {
				ws.flushBatch()
			}
			ws.mu.Unlock()

		case <-ws.done:
			ws.mu.Lock()
			ws.flushBatch()
			ws.writer.Flush()
			ws.file.Sync()
			ws.file.Close()
			ws.mu.Unlock()
			return
		}
	}
}

func (ws *WALShard) flushBatch() {
	if len(ws.batch) == 0 {
		return
	}

	for _, entry := range ws.batch {
		if err := ws.encoder.Encode(entry); err != nil {
			// Handle error (could implement retry logic)
			fmt.Printf("WAL write error: %v\n", err)
		}
	}

	if err := ws.writer.Flush(); err != nil {
		fmt.Printf("WAL flush error: %v\n", err)
	}

	// Optional: fsync for stronger durability (with performance tradeoff)
	// if err := ws.file.Sync(); err != nil {
	//     fmt.Printf("WAL sync error: %v\n", err)
	// }

	ws.batch = ws.batch[:0] // Clear batch
}

// Close cleanly shuts down the WAL system
func (ws *WALSystem) Close() error {
	for _, shard := range ws.shards {
		close(shard.done)
	}
	return nil
}

// Recover rebuilds the in-memory map from WAL files
func (ws *WALSystem) Recover(walPath string) error {
	for i := 0; i < ws.shardCount; i++ {
		shardPath := fmt.Sprintf("%s.%d", walPath, i)
		if err := ws.recoverShard(shardPath); err != nil {
			return err
		}
	}
	return nil
}

func (ws *WALSystem) recoverShard(path string) error {
	file, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // New WAL, nothing to recover
		}
		return err
	}
	defer file.Close()

	decoder := gob.NewDecoder(file)
	for {
		var entry LogEntry
		if err := decoder.Decode(&entry); err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		// Verify checksum for data integrity
		if entry.Op == 'U' && entry.Checksum != calculateChecksum(entry.Value) {
			return fmt.Errorf("checksum mismatch for key %s", entry.Key)
		}

		switch entry.Op {
		case 'U':
			ws.memMap.Store(entry.Key, entry.Value)
		case 'D':
			ws.memMap.Delete(entry.Key)
		}
	}

	return nil
}

func calculateChecksum(data []byte) uint32 {
	h := fnv.New32a()
	h.Write(data)
	return h.Sum32()
}

// Stats returns current WAL statistics
func (ws *WALSystem) Stats() (writeCount, flushCount uint64) {
	return atomic.LoadUint64(&ws.writeCount), atomic.LoadUint64(&ws.flushCount)
}

func main() {
	// Example usage
	wal, err := NewWALSystem(1, 1000, "wal/mydb.db")
	if err != nil {
		panic(err)
	}
	defer wal.Close()

	// Recover from existing WAL (if any)
	if err := wal.Recover("wal/mydb.db"); err != nil {
		panic(err)
	}

	// Simple benchmark
	start := time.Now()
	var wg sync.WaitGroup
	for i := 0; i < 1000000; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			key := fmt.Sprintf("key%d", i)
			value := []byte(fmt.Sprintf("value%d", i))
			if err := wal.Put(key, value); err != nil {
				fmt.Printf("Error writing: %v\n", err)
			}
		}(i)
	}
	wg.Wait()

	fmt.Printf("Processed 1M writes in %v\n", time.Since(start))
	writes, flushes := wal.Stats()
	fmt.Printf("Total writes: %d, flushes: %d\n", writes, flushes)
}
