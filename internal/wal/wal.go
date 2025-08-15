package wal

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/helper"
	"github.com/parashmaity/fleare/internal/logger"
	"google.golang.org/protobuf/proto"
)

// ---------------------------------------------
// Config & constants
// ---------------------------------------------

const (
	defaultBufSize       = 4 << 20 // 4 MiB bufio buffer
	defaultBatchBytes    = 1 << 20 // 1 MiB group-commit target
	defaultFlushInterval = 5 * time.Millisecond
)

// SyncPolicy controls how often fsync/fdatasync is called
// High throughput needs group commit; durability latency is a tradeoff.

type SyncPolicy int

const (
	SyncNever       SyncPolicy = iota // rely on OS flush (fastest, least durable)
	SyncEveryBatch                    // fdatasync once per batch
	SyncEveryNBatch                   // fdatasync every N batches (see SyncEveryN)
)

// Options for the WAL system

type Options struct {
	ShardCount     int
	WALPath        string
	BufSize        int           // bufio writer size per shard
	BatchBytes     int           // target bytes per batch before flush
	FlushInterval  time.Duration // max time to wait before flushing even if BatchBytes not reached
	WriteChanSize  int           // per-shard channel capacity
	SyncPolicy     SyncPolicy
	SyncEveryN     int   // used when SyncPolicy==SyncEveryNBatch
	SegmentMaxSize int64 // rotate when segment size exceeded (optional; 0=disabled)
}

func (o *Options) withDefaults() *Options {
	n := *o
	if n.BufSize <= 0 {
		n.BufSize = defaultBufSize
	}
	if n.BatchBytes <= 0 {
		n.BatchBytes = defaultBatchBytes
	}
	if n.FlushInterval <= 0 {
		n.FlushInterval = defaultFlushInterval
	}
	if n.WriteChanSize <= 0 {
		n.WriteChanSize = 65536
	}
	if n.ShardCount <= 0 {
		n.ShardCount = max(2, runtime.GOMAXPROCS(0))
	}
	if n.SyncEveryN <= 0 {
		n.SyncEveryN = 10
	}
	if n.WALPath == "" {
		n.WALPath = "."
	}
	return &n
}

// ---------------------------------------------
// WAL implementation
// ---------------------------------------------

type WALSystem struct {
	shards     map[string]*walShard
	chash      *helper.ConsistentHash
	writeCount uint64
	entryPool  sync.Pool
	closed     atomic.Bool
}

func NewWALSystem(opts Options) (*WALSystem, error) {
	o := opts.withDefaults()
	if err := os.MkdirAll(o.WALPath, 0o755); err != nil {
		return nil, fmt.Errorf("create WAL dir: %w", err)
	}

	ws := &WALSystem{
		shards:    make(map[string]*walShard, o.ShardCount),
		chash:     helper.NewConsistentHash(),
		entryPool: sync.Pool{New: func() any { return new(comm.WALEntry) }},
	}

	for i := 0; i < o.ShardCount; i++ {
		id := strconv.Itoa(i)
		segPath := filepath.Join(o.WALPath, fmt.Sprintf("data-file-%s.db", id))
		sh, err := newWalShard(id, segPath, o)
		if err != nil {
			return nil, err
		}
		ws.shards[id] = sh
		ws.chash.AddNode(id)
	}
	return ws, nil
}

func (ws *WALSystem) Close() error {
	if ws.closed.Swap(true) {
		return nil
	}
	var first error
	for _, s := range ws.shards {
		if s == nil {
			continue
		}
		if err := s.Close(); err != nil && first == nil {
			first = err
		}
	}
	return first
}

func (ws *WALSystem) GetShard(name string) Shard { return ws.shards[name] }

func (ws *WALSystem) GetShardByKey(key string) Shard {
	node := ws.chash.GetNode(key)
	return ws.shards[node]
}

func (ws *WALSystem) Put(key string, obj *comm.Object) error {
	if ws.closed.Load() {
		return errors.New("wal: closed")
	}
	atomic.AddUint64(&ws.writeCount, 1)
	entry := ws.entryPool.Get().(*comm.WALEntry)
	entry.Op = comm.WALEntry_UPDATE
	entry.Key = key
	entry.Object = obj
	// Checksum is calculated on encoded bytes inside shard goroutine (zero-alloc path)
	return ws.enqueue(key, entry)
}

func (ws *WALSystem) Delete(key string) error {
	if ws.closed.Load() {
		return errors.New("wal: closed")
	}
	atomic.AddUint64(&ws.writeCount, 1)
	entry := ws.entryPool.Get().(*comm.WALEntry)
	entry.Op = comm.WALEntry_DELETE
	entry.Key = key
	entry.Object = nil
	return ws.enqueue(key, entry)
}

func (ws *WALSystem) enqueue(key string, e *comm.WALEntry) error {
	sh := ws.GetShardByKey(key).(*walShard)
	select {
	case sh.ch <- e:
		return nil
	default:
		// backpressure: block with timeout proportional to FlushInterval
		select {
		case sh.ch <- e:
			return nil
		case <-time.After(sh.opts.FlushInterval * 4):
			return errors.New("wal shard channel full: backpressure")
		}
	}
}

func (ws *WALSystem) Stats() uint64 { return atomic.LoadUint64(&ws.writeCount) }

// ---------------------------------------------
// Shard
// ---------------------------------------------

type walShard struct {
	id     string
	file   *os.File
	bufw   *bufio.Writer
	ch     chan *comm.WALEntry
	opts   *Options
	closed atomic.Bool

	// batching
	batchBuf        bytes.Buffer
	batchBytes      int
	batchCount      int
	syncEveryTicker *time.Ticker
	flushTimer      *time.Timer

	// pools to reduce allocations
	encBufPool sync.Pool // []byte buffers for MarshalAppend path

	wg sync.WaitGroup
}

func newWalShard(id, path string, o *Options) (*walShard, error) {
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	sh := &walShard{
		id:         id,
		file:       f,
		bufw:       bufio.NewWriterSize(f, o.BufSize),
		ch:         make(chan *comm.WALEntry, o.WriteChanSize),
		opts:       o,
		encBufPool: sync.Pool{New: func() any { b := make([]byte, 0, 4<<10); return &b }},
	}

	sh.flushTimer = time.NewTimer(o.FlushInterval)
	if !sh.flushTimer.Stop() {
		<-sh.flushTimer.C
	}

	sh.wg.Add(1)
	go sh.run()
	return sh, nil
}

func (s *walShard) Close() error {
	if s.closed.Swap(true) {
		return nil
	}
	close(s.ch)
	s.wg.Wait()
	if err := s.bufw.Flush(); err != nil {
		return err
	}
	if s.opts.SyncPolicy != SyncNever {
		if err := s.file.Sync(); err != nil {
			return err
		}
	}
	return s.file.Close()
}

// Record encoding layout: [u32 length][u32 crc32c][protobuf bytes]
// This makes recovery simple and checksum robust.

var crcTab = crc32.MakeTable(crc32.Castagnoli)

func (s *walShard) run() {
	defer s.wg.Done()
	batchDeadline := time.Now().Add(s.opts.FlushInterval)

	for {
		var e *comm.WALEntry
		var ok bool

		// fast path: drain as much as possible up to BatchBytes or deadline
		select {
		case e, ok = <-s.ch:
			if !ok {
				s.flush(true)
				return
			}
			s.encodeAppend(e)
			// recycle entry back to pool
			*e = comm.WALEntry{}
			// can't access ws.entryPool here; keep minimal: GC will handle or inject external pool via callback if needed
			// NOTE: caller owns a pool; if needed, make WALSystem expose a PutEntry method
		case <-time.After(time.Until(batchDeadline)):
			// timeout triggers flush
		}

		// inner drain loop to coalesce many entries cheaply
		for s.batchBytes < s.opts.BatchBytes {
			select {
			case e, ok = <-s.ch:
				if !ok {
					s.flush(true)
					return
				}
				s.encodeAppend(e)
				*e = comm.WALEntry{}
			default:
				goto FLUSH_CHECK
			}
		}

	FLUSH_CHECK:
		if s.batchBytes >= s.opts.BatchBytes || time.Now().After(batchDeadline) {
			s.flush(false)
			batchDeadline = time.Now().Add(s.opts.FlushInterval)
		}
	}
}

func (s *walShard) encodeAppend(e *comm.WALEntry) {
	// marshal protobuf into pooled []byte via MarshalAppend to reduce allocs
	bufPtr := s.encBufPool.Get().(*[]byte)
	b := (*bufPtr)[:0]
	// MarshalOptions with AllowPartial speeds slightly; use standard for safety here
	var mo proto.MarshalOptions
	var data []byte
	var err error
	data, err = mo.MarshalAppend(b, e)
	if err != nil {
		logger.Error("wal: marshal", err, map[string]any{"key": e.Key})
		s.encBufPool.Put(bufPtr)
		return
	}

	crc := crc32.Checksum(data, crcTab)
	// record = len(4) + crc(4) + data
	recLen := 4 + len(data)
	// header length = u32(recLen)
	var header [8]byte
	binary.BigEndian.PutUint32(header[0:4], uint32(recLen))
	binary.BigEndian.PutUint32(header[4:8], crc)

	s.batchBuf.Write(header[:])
	s.batchBuf.Write(data)
	s.batchBytes += 8 + len(data)
	s.batchCount++

	// reuse the buffer for next time
	*bufPtr = data[:0]
	s.encBufPool.Put(bufPtr)
}

func (s *walShard) flush(forceSync bool) {
	if s.batchBytes == 0 {
		return
	}
	if _, err := s.bufw.Write(s.batchBuf.Bytes()); err != nil {
		logger.Error("wal: write batch", err, nil)
	}
	if err := s.bufw.Flush(); err != nil {
		logger.Error("wal: flush", err, nil)
	}

	needSync := false
	switch s.opts.SyncPolicy {
	case SyncEveryBatch:
		needSync = true
	case SyncEveryNBatch:
		needSync = (s.batchCount % s.opts.SyncEveryN) == 0
	}
	if forceSync {
		needSync = true
	}
	if needSync {
		if err := s.file.Sync(); err != nil {
			logger.Error("wal: sync", err, nil)
		}
	}

	// reset
	s.batchBuf.Reset()
	s.batchBytes = 0
}

// ---------------------------------------------
// Recovery (tolerant to torn tail)
// ---------------------------------------------

func (ws *WALSystem) Recover(walPath string, callback func(key string, obj *comm.Object, isDeleted bool)) error {
	// iterate shards by count known in this instance. If unknown, scan directory by prefix.
	for id := range ws.shards {
		p := filepath.Join(walPath, fmt.Sprintf("data-file-%s.db", id))
		if err := recoverShard(p, callback); err != nil {
			return err
		}
	}
	return nil
}

func recoverShard(path string, cb func(key string, obj *comm.Object, isDeleted bool)) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer f.Close()

	r := bufio.NewReaderSize(f, defaultBufSize)
	for {
		var hdr [8]byte
		if _, err := io.ReadFull(r, hdr[:]); err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			if errors.Is(err, io.ErrUnexpectedEOF) {
				break
			} // tolerate torn tail
			return err
		}
		recLen := binary.BigEndian.Uint32(hdr[0:4])
		expectedCRC := binary.BigEndian.Uint32(hdr[4:8])
		if recLen < 4 {
			return fmt.Errorf("wal: bad recLen %d", recLen)
		}

		data := make([]byte, recLen-4) // recLen includes crc(4)+data
		if _, err := io.ReadFull(r, data); err != nil {
			if errors.Is(err, io.ErrUnexpectedEOF) {
				break
			}
			return err
		}
		if crc32.Checksum(data, crcTab) != expectedCRC {
			return fmt.Errorf("wal: checksum mismatch")
		}

		var e comm.WALEntry
		if err := proto.Unmarshal(data, &e); err != nil {
			return err
		}
		switch e.Op {
		case comm.WALEntry_UPDATE:
			cb(e.Key, e.Object, false)
		case comm.WALEntry_DELETE:
			cb(e.Key, nil, true)
		}
	}
	return nil
}

// ---------------------------------------------
// Helpers
// ---------------------------------------------

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
