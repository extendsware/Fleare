package store

// import (
// 	"context"
// 	"sync"
// 	"time"

// 	"github.com/parashmaity/fleare/internal/logger"
// )

// // TTLManager manages the background cleanup of expired keys
// type TTLManager struct {
// 	memories        []IMemory
// 	cleanupInterval time.Duration
// 	ctx             context.Context
// 	cancel          context.CancelFunc
// 	wg              sync.WaitGroup
// 	mu              sync.RWMutex
// }

// // NewTTLManager creates a new TTL manager with the specified cleanup interval
// func NewTTLManager(cleanupInterval time.Duration) *TTLManager {
// 	ctx, cancel := context.WithCancel(context.Background())
// 	return &TTLManager{
// 		memories:        make([]IMemory, 0),
// 		cleanupInterval: cleanupInterval,
// 		ctx:             ctx,
// 		cancel:          cancel,
// 	}
// }

// // RegisterMemory registers a memory instance for TTL cleanup
// func (tm *TTLManager) RegisterMemory(memory IMemory) {
// 	tm.mu.Lock()
// 	defer tm.mu.Unlock()
// 	tm.memories = append(tm.memories, memory)
// }

// // Start begins the background cleanup process
// func (tm *TTLManager) Start() {
// 	tm.wg.Add(1)
// 	go tm.cleanupLoop()
// 	logger.Info("TTL Manager started with cleanup interval: %v", tm.cleanupInterval)
// }

// // Stop stops the background cleanup process
// func (tm *TTLManager) Stop() {
// 	tm.cancel()
// 	tm.wg.Wait()
// 	logger.Info("TTL Manager stopped")
// }

// // cleanupLoop runs the periodic cleanup of expired keys
// func (tm *TTLManager) cleanupLoop() {
// 	defer tm.wg.Done()

// 	ticker := time.NewTicker(tm.cleanupInterval)
// 	defer ticker.Stop()

// 	for {
// 		select {
// 		case <-tm.ctx.Done():
// 			return
// 		case <-ticker.C:
// 			tm.performCleanup()
// 		}
// 	}
// }

// // performCleanup removes expired keys from all registered memories
// func (tm *TTLManager) performCleanup() {
// 	tm.mu.RLock()
// 	memories := make([]IMemory, len(tm.memories))
// 	copy(memories, tm.memories)
// 	tm.mu.RUnlock()

// 	totalCleaned := int32(0)
// 	for _, memory := range memories {
// 		cleaned := memory.CleanupExpired()
// 		totalCleaned += cleaned
// 	}

// 	if totalCleaned > 0 {
// 		logger.Debug("TTL cleanup removed %d expired keys", totalCleaned)
// 	}
// }

// // ForceCleanup immediately performs cleanup on all registered memories
// func (tm *TTLManager) ForceCleanup() int32 {
// 	tm.mu.RLock()
// 	defer tm.mu.RUnlock()

// 	totalCleaned := int32(0)
// 	for _, memory := range tm.memories {
// 		cleaned := memory.CleanupExpired()
// 		totalCleaned += cleaned
// 	}

// 	logger.Debug("Forced TTL cleanup removed %d expired keys", totalCleaned)
// 	return totalCleaned
// }

// // GetCleanupInterval returns the current cleanup interval
// func (tm *TTLManager) GetCleanupInterval() time.Duration {
// 	return tm.cleanupInterval
// }

// // SetCleanupInterval updates the cleanup interval
// func (tm *TTLManager) SetCleanupInterval(interval time.Duration) {
// 	tm.cleanupInterval = interval
// 	logger.Info("TTL cleanup interval updated to: %v", interval)
// }
