//go:build linux
// +build linux

package multiplexer

import (
	"fmt"
	"time"

	"golang.org/x/sys/unix"
)

const (
	EVFILT_READ  = 0x1
	EVFILT_WRITE = 0x4
	EV_ADD       = 0x1
	EV_CLEAR     = 0x20
	EV_DELETE    = 0x2
	EV_DISABLE   = 0x8
	EV_ENABLE    = 0x4
	EV_EOF       = 0x8000
)

type Epoll struct {
	fd     int
	events []unix.EpollEvent
	// changes tracks the event changes to be applied to kqueue
	// changes []unix.EpollEvent
}

func NewPoller(maxEvents int) (*Epoll, error) {
	fd, err := unix.EpollCreate1(0)
	if err != nil {
		return nil, fmt.Errorf("epoll create: %w", err)
	}
	return &Epoll{
		fd:     fd,
		events: make([]unix.EpollEvent, maxEvents),
		// changes: make([]unix.EpollEvent, MaxEvents),
	}, nil
}

func (e *Epoll) Register(fd int, filter int16) error {
	var ev unix.EpollEvent
	ev.Fd = int32(fd)

	if filter == EVFILT_READ {
		ev.Events = unix.EPOLLIN | unix.EPOLLET
	} else if filter == EVFILT_WRITE {
		ev.Events = unix.EPOLLOUT | unix.EPOLLET
	}

	err := unix.EpollCtl(e.fd, unix.EPOLL_CTL_ADD, fd, &ev)
	if err != nil {
		return err
	}
	return nil
}

func (e *Epoll) Unregister(fd int, filter int16) error {
	err := unix.EpollCtl(e.fd, unix.EPOLL_CTL_DEL, fd, nil)
	if err != nil {
		return err
	}
	return nil
}

func (e *Epoll) Poll(timeout time.Duration, callback func(fd int, filter int32, flags int32) error) error {
	timeoutMillis := int(timeout / time.Millisecond)
	numEvents, err := unix.EpollWait(e.fd, e.events, timeoutMillis)
	if err != nil {
		if err == unix.EINTR {
			return nil
		}
		return fmt.Errorf("epoll wait failed: %v", err)
	}
	if numEvents == 0 {
		return nil
	}

	for _, v := range e.events[:numEvents] {
		// EpollEvent doesn't have Flags field, and unix doesn't define EV_EOF or EV_ERROR
		// Instead, check for EPOLLERR, EPOLLHUP which indicate errors
		if (v.Events&unix.EPOLLERR) != 0 || (v.Events&unix.EPOLLHUP) != 0 {
			continue
		}
		err := callback(int(v.Fd), int32(v.Events), v.Pad)
		if err != nil {
			break
		}
	}
	return nil
}

func (e *Epoll) Close() error {
	return unix.Close(e.fd)
}
