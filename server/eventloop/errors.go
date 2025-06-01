package eventloop

import "errors"

var (
	ErrAuthFailed       = errors.New("AuthenticationFailed")
	ErrAcceptConnection = errors.New("ConnectionFailed")
	ErrReadData         = errors.New("failed to read data")
	ErrWriteData        = errors.New("failed to write data")

	ErrAborted = errors.New("server received ABORT command")

	ErrRequestTooLarge = errors.New("request too large")
	ErrIdleTimeout     = errors.New("connection idle timeout")
	ErrorClosed        = errors.New("connection closed")

	ErrMaxClientsReached = errors.New("maximum number of clients reached")
	ErrIOThreadNotFound  = errors.New("io-thread not found")

	ErrInvalidFileDescriptor = errors.New("invalid file descriptor")
	ErrTimeout               = errors.New("operation timed out")
	ErrClosed                = errors.New("multiplexer is closed")
)
