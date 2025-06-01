package eventloop

import (
	"sync"
	"sync/atomic"
)

type IOThreadManager struct {
	connectedClients sync.Map
	numIOThreads     atomic.Uint32
	mu               sync.Mutex
}

func NewIOThreadManager() *IOThreadManager {
	return &IOThreadManager{}
}

func (m *IOThreadManager) Register(ioThread *IOThread) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.IOThreadCount() >= uint32(100000) {
		return ErrMaxClientsReached
	}

	m.connectedClients.Store(ioThread.ClientID, ioThread)
	m.numIOThreads.Add(1)
	return nil
}

func (m *IOThreadManager) Unregister(id string) error {
	if client, loaded := m.connectedClients.LoadAndDelete(id); loaded {
		w := client.(*IOThread)
		if err := w.Stop(); err != nil {
			return err
		}
	} else {
		return ErrIOThreadNotFound
	}

	m.numIOThreads.Add(^uint32(0))
	return nil
}

func (m *IOThreadManager) GetThread(clientID string) (*IOThread, error) {
	if client, loaded := m.connectedClients.Load(clientID); loaded {
		return client.(*IOThread), nil
	}
	return nil, ErrIOThreadNotFound
}

func (m *IOThreadManager) IOThreadCount() uint32 {
	return m.numIOThreads.Load()
}

func (m *IOThreadManager) StopAll() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.connectedClients.Range(func(key, value interface{}) bool {
		ioThread := value.(*IOThread)
		ioThread.Stop()
		return true
	})

	m.numIOThreads.Store(0)
}
