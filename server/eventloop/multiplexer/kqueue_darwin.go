//go:build darwin || freebsd || netbsd || openbsd
// +build darwin freebsd netbsd openbsd

package multiplexer

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

const (
	EVFILT_READ  = -0x1
	EVFILT_WRITE = -0x2
	EV_ADD       = 0x1
	EV_CLEAR     = 0x20
	EV_DELETE    = 0x2
	EV_DISABLE   = 0x8
	EV_ENABLE    = 0x4
	EV_EOF       = 0x8000
)

type Kqueue struct {
	// kq is the kqueue file descriptor used for event notification
	kq int
	// events stores the array of events returned from kqueue
	events []unix.Kevent_t
	// changes tracks the event changes to be applied to kqueue
	// changes []unix.Kevent_t
	// closed indicates if the kqueue has been closed
	closed bool
	// mu protects concurrent access to kqueue operations
	// mu sync.RWMutex
}

func NewPoller(maxEvents int) (*Kqueue, error) {
	kq, err := unix.Kqueue()
	if err != nil {
		return nil, fmt.Errorf("failed to create kqueue: %v", err)
	}
	return &Kqueue{
		kq:     kq,
		events: make([]unix.Kevent_t, maxEvents),
		// changes: make([]unix.Kevent_t, MaxEvents),
	}, nil
}

func (k *Kqueue) Register(fd int, filter int16) error {
	// k.mu.Lock()
	// defer k.mu.Unlock()

	if k.closed {
		return fmt.Errorf("kqueue is closed")
	}

	event := []unix.Kevent_t{
		{
			Ident:  uint64(fd),
			Filter: filter,
			Flags:  EV_ADD | EV_ENABLE,
			Fflags: 0,
			Data:   0,
			Udata:  nil,
		},
	}

	_, err := unix.Kevent(k.kq, event, nil, nil)
	if err != nil {
		return fmt.Errorf("failed to register fd %d with filter %d: %v", fd, filter, err)
	}

	// k.changes = append(k.changes, event)
	return nil
}

func (k *Kqueue) Unregister(fd int, filter int16) error {
	// k.mu.Lock()
	// defer k.mu.Unlock()

	if k.closed {
		return nil // No error if already closed
	}

	event := []unix.Kevent_t{
		{Ident: uint64(fd),
			Filter: filter,
			Flags:  EV_DELETE,
			Fflags: 0,
			Data:   0,
			Udata:  nil,
		},
	}
	// k.changes = append(k.changes, event)
	_, err := unix.Kevent(k.kq, event, nil, nil)
	if err != nil {
		// If the event doesn't exist (ENOENT) or the descriptor is bad (EBADF),
		// it's not a fatal error during cleanup
		if err == unix.ENOENT || err == unix.EBADF {
			return nil
		}
		return fmt.Errorf("failed to unregister fd %d with filter %d: %v", fd, filter, err)
	}
	return nil
}

func (k *Kqueue) Poll(timeout time.Duration, callback func(fd int, filter int32, flags int32) error) error {
	// k.mu.RLock()
	// defer k.mu.RUnlock()

	if k.closed {
		return fmt.Errorf("kqueue is closed")
	}

	timeoutPtr := &unix.Timespec{
		Sec:  int64(timeout / time.Second),
		Nsec: int64(timeout % time.Second),
	}
	numEvents, err := unix.Kevent(k.kq, nil, k.events, timeoutPtr)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
		return fmt.Errorf("kevent error: %v", err)
	}
	if numEvents == 0 {
		return nil
	}
	for _, v := range k.events[:numEvents] {
		if (v.Flags&unix.EV_EOF) != 0 || (v.Flags&unix.EV_ERROR) != 0 {
			continue
		}
		err := callback(int(v.Ident), int32(v.Filter), int32(v.Flags))
		if err != nil {
			break
		}
	}
	return nil
}

func (k *Kqueue) Close() error {
	// k.mu.Lock()
	// defer k.mu.Unlock()

	if k.closed {
		return nil
	}
	k.closed = true
	return unix.Close(k.kq)
}
