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

	unix.Kevent(k.kq, event, nil, nil)

	// k.changes = append(k.changes, event)
	return nil
}

func (k *Kqueue) Unregister(fd int, filter int16) error {
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
	unix.Kevent(k.kq, event, nil, nil)
	return nil
}

func (k *Kqueue) Poll(timeout time.Duration, callback func(fd int, filter int32, flags int32) error) error {
	timeoutPtr := &unix.Timespec{
		Sec:  int64(timeout / time.Second),
		Nsec: int64(timeout % time.Second),
	}
	numEvents, err := unix.Kevent(k.kq, nil, k.events, timeoutPtr)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
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
	return unix.Close(k.kq)
}
