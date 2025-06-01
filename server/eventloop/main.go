package eventloop

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/parashmaity/fleare/config"
	iomanager "github.com/parashmaity/fleare/internal/io-manager"
	logger "github.com/parashmaity/fleare/internal/logger"
	"github.com/parashmaity/fleare/internal/utils"
	"github.com/parashmaity/fleare/server/eventloop/multiplexer"

	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
)

var (
	serverErrCh = make(chan error, 2)
)

type Server struct {
	Host            string
	Port            int
	serverFD        int
	ioThreadManager *IOThreadManager
	activeThreads   sync.WaitGroup
	IOServer        IOServer
	Opts            *Options
	eventLoop       multiplexer.EventLoop
}

// exponentialBackoff implements a simple exponential backoff strategy
type exponentialBackoff struct {
	initialDelay time.Duration
	maxDelay     time.Duration
	factor       float64
	currentDelay time.Duration
}

func NewServer(ioThreadManager *IOThreadManager, ioserver IOServer, opts *Options) *Server {
	return &Server{
		Host:            config.GetConfig().Server.Host,
		Port:            config.GetConfig().Server.Port,
		ioThreadManager: ioThreadManager,
		IOServer:        ioserver,
		Opts:            opts,
	}
}

func Start(IServer IOServer, threadManager *iomanager.ThreadManager, opts *Options) {

	ioThreadManager := NewIOThreadManager()
	srv := NewServer(ioThreadManager, IServer, opts)

	threadManager.Go("run-server", func(ctx context.Context) {
		runServer(ctx, srv, serverErrCh)
		close(serverErrCh)
	})

	// Process server errors
	for err := range serverErrCh {
		if err != nil && !errors.Is(err, context.Canceled) {
			logger.Error("Server error", err, nil)
			if errors.Is(err, ErrAborted) {
				threadManager.CancelAll()
			}
		}
	}
	IServer.OnStop()
	logger.Info("Stoped", nil)
}

func runServer(ctx context.Context, srv *Server, errCh chan<- error) {
	if err := srv.StartEventLoop(ctx); err != nil {
		switch {
		case errors.Is(err, context.Canceled):
			logger.Debug(fmt.Sprintf("%T was canceled", srv), nil)
		case errors.Is(err, ErrAborted):
			logger.Debug(fmt.Sprintf("%T received abort command", srv), nil)
		default:
			logger.Error(fmt.Sprintf("%T error", srv), err, nil)
		}
		errCh <- err
	} else {
		logger.Debug("bye.", nil)
	}
}

func (s *Server) StartEventLoop(ctx context.Context) (err error) {
	if err = s.BindAndListen(); err != nil {
		logger.Error("failed to bind server", err, nil)
		return err
	}

	defer releasePort(s.serverFD)

	eloop, err := multiplexer.New(config.GetConfig().Misc.MaxConnections)
	if err != nil {
		logger.Error("failed to create event loop", err, nil)
		return err
	}

	s.eventLoop = eloop

	s.eventLoop.Register(s.serverFD, multiplexer.EVFILT_READ)

	errChan := make(chan error, 1)
	acceptWg := &sync.WaitGroup{}

	acceptWg.Add(1)
	go func(wg *sync.WaitGroup) {
		defer wg.Done()
		if err := s.AcceptNewConnection(ctx); err != nil && !errors.Is(err, context.Canceled) {
			errChan <- fmt.Errorf("failed to accept connections %w", err)
		}
	}(acceptWg)

	// Wait for accept goroutine to finish
	acceptWg.Wait()

	select {
	case <-ctx.Done():
		logger.Console("Initiating shutdown", zerolog.WarnLevel, nil)
	case err = <-errChan:
		logger.Error("Error while accepting connections, initiating shutdown", err, nil)
	}

	// Wait for all client connections to be properly closed
	waitCh := make(chan struct{})
	go func() {
		s.activeThreads.Wait()
		close(waitCh)
	}()

	// Add a timeout for graceful shutdown
	select {
	case <-waitCh:
		logger.Console("All client connections closed", zerolog.InfoLevel, nil)
	case <-time.After(1 * time.Second):
		logger.Console("Shutdown timeout reached, some connections may not have closed properly", zerolog.WarnLevel, nil)
	}
	logger.Info("Exiting gracefully", nil)
	return err
}

func (s *Server) BindAndListen() error {
	serverFD, socketErr := unix.Socket(unix.AF_INET, unix.SOCK_STREAM, 0)
	if socketErr != nil {
		return fmt.Errorf("failed to create socket: %w", socketErr)
	}

	// Close the socket on exit if an error occurs
	var err error
	defer func() {
		if err != nil {
			if closeErr := unix.Close(serverFD); closeErr != nil {
				// Wrap the close error with the original bind/listen error
				logger.Error("Error occurred", err, map[string]any{"additionally": "failed to close socket", "close-err": closeErr})
			} else {
				logger.Error("Error occurred", err, nil)
			}
		}
	}()

	if err = unix.SetsockoptInt(serverFD, unix.SOL_SOCKET, unix.SO_REUSEADDR, 1); err != nil {
		return fmt.Errorf("failed to set SO_REUSEADDR: %w", err)
	}

	if err = unix.SetNonblock(serverFD, true); err != nil {
		return fmt.Errorf("failed to set socket to non-blocking: %w", err)
	}

	ip4 := s.getOutboundIP()
	if ip4 == nil {
		return fmt.Errorf("invalid IP address: %s", s.Host)
	}

	logger.Console("-> Fleare server started", zerolog.InfoLevel, map[string]any{
		"IP-Address:": ip4.String(),
		"Port":        s.Port,
	})

	s.IOServer.OnStart(ip4.String(), s.Port)

	sockAddr := &unix.SockaddrInet4{
		Port: s.Port,
		Addr: [4]byte{ip4[0], ip4[1], ip4[2], ip4[3]},
	}
	if err = unix.Bind(serverFD, sockAddr); err != nil {
		return fmt.Errorf("failed to bind socket: %w", err)
	}

	if err = unix.Listen(serverFD, unix.SOMAXCONN); err != nil {
		return fmt.Errorf("failed to listen on socket: %w", err)
	}

	s.serverFD = serverFD
	return nil
}

func (s *Server) getOutboundIP() net.IP {
	if s.Host == "0.0.0.0" {
		// Try to get the outbound IP using Google's DNS server
		conn, err := net.DialTimeout("udp", "8.8.8.8:80", time.Second*2) // Added timeout to prevent long waiting
		if err == nil {
			defer conn.Close()
			localAddr := conn.LocalAddr().(*net.UDPAddr)
			return localAddr.IP
		}

		// If internet is not available, fall back to localhost
		logger.Warn("Cannot connect to external network, falling back to localhost", map[string]any{
			"error": err.Error(),
		})
	}

	return net.ParseIP("127.0.0.1")
}

func (s *Server) AcceptNewConnection(ctx context.Context) error {
	// Use epoll or poll for more efficient I/O multiplexing in production
	pollTimeout := 100 * time.Millisecond
	backoffStrategy := &exponentialBackoff{
		initialDelay: 5 * time.Millisecond,
		maxDelay:     100 * time.Millisecond,
		factor:       1.5,
		currentDelay: 5 * time.Millisecond,
	}

	ipAddrBuffer := new(strings.Builder)

	for {
		select {
		case <-ctx.Done():
			logger.Info("No new connections will be accepted", nil)
			return ctx.Err()
		default:

			eventHandler := func(fd int, filter int32, flags int32) error {

				if fd == s.serverFD && filter == multiplexer.EVFILT_READ {
					clientFD, clientAddr, acceptErr := unix.Accept(s.serverFD)

					if acceptErr != nil {
						if errors.Is(acceptErr, unix.EAGAIN) || errors.Is(acceptErr, unix.EWOULDBLOCK) {
							// No more connections to accept
							backoffStrategy.reset()
							return acceptErr
						} else if errors.Is(acceptErr, unix.EINTR) {
							return nil // Retry this accept
						} else if errors.Is(acceptErr, unix.EMFILE) || errors.Is(acceptErr, unix.ENFILE) {
							logger.Error("Too many open files, backing off", acceptErr, nil)
							time.Sleep(pollTimeout)
							return fmt.Errorf("Too many open files, backing off %v", acceptErr)
						} else {
							backoffStrategy.reset()
							return fmt.Errorf("error accepting connection: %w", acceptErr)
						}
					}

					// Efficiently format IP address - reuse builder
					ipAddrBuffer.Reset()
					utils.FormatClientAddr(ipAddrBuffer, clientAddr)

					// Generate a unique client ID
					clientID := GenerateUniqueClientID(clientFD, ipAddrBuffer.String())
					logger.Debug("Generated client ID", map[string]any{"client_id": clientID, "fd": clientFD})

					// Authenticate the client
					connect := s.ConnectAndAuthenticate(clientFD, clientID)
					if !connect {
						unix.Close(clientFD)
						return nil
					}

					// Batch socket options to reduce syscalls
					// Set TCP keepalive and non-blocking mode together
					if err := unix.SetsockoptInt(clientFD, unix.SOL_SOCKET, unix.SO_KEEPALIVE, 1); err != nil {
						logger.Error("Failed to set SO_KEEPALIVE", err, nil)
					}
					if err := unix.SetNonblock(clientFD, true); err != nil {
						logger.Error("Failed to set client socket to non-blocking", err, nil)
						unix.Close(clientFD)
						return nil
					}

					// register read fd operations
					s.eventLoop.Register(clientFD, multiplexer.EVFILT_READ)

					// Try to get a thread from the IO thread manager first
					var thread *IOThread
					var threadErr error

					if s.ioThreadManager != nil {
						// Call the appropriate method on the IOThreadManager
						// Assuming there's an Acquire or similar method instead of GetThread
						thread, _ = s.ioThreadManager.GetThread(clientID)
					}

					// If no thread was available from the manager, create a new one
					if thread == nil {
						thread, threadErr = NewIOThread(clientFD, clientID, s.IOServer)
						if threadErr != nil {
							logger.Error("Failed to create io-thread", threadErr, map[string]any{"client-id": clientID})
							unix.Close(clientFD)
							return nil
						}
					}

					// Track this active connection and start processing
					s.activeThreads.Add(1)
					go func(t *IOThread) {
						if err := s.ioThreadManager.Register(t); err != nil {
							logger.Error("Failed to register io-thread", err, map[string]any{"client-id": clientID})
							unix.Close(clientFD)
							s.activeThreads.Done()
						} else {
							s.startConnThread(ctx, t)
						}
					}(thread)

				}

				return nil
			}

			err := s.eventLoop.Poll(10*time.Millisecond, eventHandler)
			if err != nil {
				return err
			}

		}
	}
}

func GenerateUniqueClientID(clientFD int, clientAddr string) string {
	clientID := fmt.Sprintf("%d-%s", clientFD, clientAddr)

	return clientID
}

func (b *exponentialBackoff) reset() {
	b.currentDelay = b.initialDelay
}

func (s *Server) startConnThread(ctx context.Context, thread *IOThread) {
	// Prepare for cleanup in a single defer block to reduce overhead
	fd := thread.IOConn.fd
	clientID := thread.ClientID

	defer func() {
		// Batch operations where possible to reduce function call overhead
		if err := s.ioThreadManager.Unregister(clientID); err != nil {
			logger.Error("Failed to unregister IO thread", err, map[string]any{"ClientID": clientID})
		}

		// Unregister both events with a single call if your eventLoop implementation supports it
		// Otherwise, keep them separate but minimal
		s.eventLoop.Unregister(fd, multiplexer.EVFILT_READ)
		s.eventLoop.Unregister(fd, multiplexer.EVFILT_WRITE)

		// Close the socket only once at the end
		if err := unix.Close(fd); err != nil && !errors.Is(err, unix.EBADF) {
			logger.Error("Failed to close client socket", err, map[string]any{"FD": fd})
		}
		thread.IOConn.Close()
		s.activeThreads.Done()
		s.IOServer.OnDisconnect(clientID)
	}()

	// Create the thread context with cancel
	threadCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Notify the server about the new connection
	s.IOServer.OnConnection(*thread.IOConn, clientID)

	// Start the thread synchronously and handle errors
	if err := thread.StartSync(threadCtx); err != nil {
		switch {
		case err == io.EOF:
			logger.Debug("client disconnected. io-thread stopped", nil)
		case errors.Is(err, context.Canceled):
			logger.Debug("io-thread canceled during server shutdown", nil)
		default:
			logger.Error("io-thread error", err, map[string]any{"FD": fd})
		}
	}
}

func releasePort(serverFD int) {
	if err := unix.Close(serverFD); err != nil {
		logger.Error("Failed to close server socket", err, nil)
	}
}
