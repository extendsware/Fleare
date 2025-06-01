package multiplexer

import "time"

type EventLoop interface {
	// Register adds a file descriptor to the event loop with the specified filter
	// fd is the file descriptor to register
	// filter specifies the type of events to monitor (e.g., read, write)
	// returns an error if registration fails
	// the file descriptor remains registered until explicitly unregistered
	Register(fd int, filter int16) error

	// Unregister removes a file descriptor from the event loop for the specified filter
	// fd is the file descriptor to unregister
	// filter specifies the type of events to stop monitoring
	// returns an error if unregistration fails
	// after unregistration, no more events will be delivered for this fd/filter combination
	Unregister(fd int, filter int16) error

	// Poll waits for I/O events on registered file descriptors
	// timeout specifies how long to wait for events (0 for immediate return, negative for indefinite wait)
	// callback is called for each event with the fd, filter type, and flags
	// returns an error if polling fails
	// blocks until timeout expires or events occur
	Poll(timeout time.Duration, callback func(fd int, filter int32, flags int32) error) error

	// Close releases resources associated with the event loop
	// should be called when the event loop is no longer needed
	// returns an error if closing fails
	// after closing, the event loop cannot be used again
	// any subsequent operations will result in undefined behavior
	Close() error
}

func New(maxEvents int) (EventLoop, error) {
	return NewPoller(maxEvents)
}
