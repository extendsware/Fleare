package iomanager

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/parashmaity/fleare/internal/logger"
)

// ThreadManager manages thread execution with context cancellation
type ThreadManager struct {
	mu       sync.RWMutex
	threads  map[string]context.CancelFunc
	wg       sync.WaitGroup
	active   int32
	ctx      context.Context
	cancelFn context.CancelFunc
}

// NewManager creates a new global thread manager.
func NewThreadManager() *ThreadManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &ThreadManager{
		threads:  make(map[string]context.CancelFunc),
		ctx:      ctx,
		cancelFn: cancel,
	}
}

// Go launches a cancellable thread with global registration.
func (m *ThreadManager) Go(id string, fn func(ctx context.Context)) {
	// Create a child context of the global context
	ctx, cancel := context.WithCancel(m.ctx)

	m.mu.Lock()
	if _, exists := m.threads[id]; exists {
		fmt.Printf("Thread %s already running\n", id)
		cancel() // Prevent context leak
		m.mu.Unlock()
		return
	}
	m.threads[id] = cancel
	m.wg.Add(1)
	atomic.AddInt32(&m.active, 1)
	m.mu.Unlock()

	go func() {
		defer func() {
			m.mu.Lock()
			delete(m.threads, id)
			atomic.AddInt32(&m.active, -1)
			m.mu.Unlock()
			m.wg.Done()
		}()
		fn(ctx)
	}()
}

// Cancel stops a specific thread.
func (m *ThreadManager) CloseThread(id string) {
	m.mu.RLock()
	cancel, ok := m.threads[id]
	m.mu.RUnlock()

	if ok {
		cancel()
	}
}

// CancelAll stops all running threads.
func (m *ThreadManager) CancelAll() {
	// Cancel the global context which will cascade to all child contexts
	m.cancelFn()

	// Also individually cancel each thread to ensure immediate notification
	m.mu.RLock()
	cancels := make([]context.CancelFunc, 0, len(m.threads))
	for _, cancel := range m.threads {
		cancels = append(cancels, cancel)
	}
	m.mu.RUnlock()

	// Execute cancellations without holding the lock
	for _, cancel := range cancels {
		cancel()
	}
}

// Wait blocks until all threads have finished.
func (m *ThreadManager) Wait() {
	m.wg.Wait()
}

func (m *ThreadManager) GetContext() context.Context {
	return m.ctx
}

func (m *ThreadManager) Shutdown(ctx context.Context) {
	logger.Info("Shutting down server...", nil)
	// Signal all connected clients to disconnect
	shutdownCtx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	// Here you would implement the notification to all clients
	// For now, we just rely on the context cancellation to propagate to all IOThreads

	select {
	case <-shutdownCtx.Done():
		logger.Info("Shutdown notification complete", nil)
	case <-ctx.Done():
		logger.Warn("Shutdown notification aborted", nil)
	}
}
