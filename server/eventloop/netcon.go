package eventloop

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"time"

	"github.com/parashmaity/fleare/internal/comm"

	"google.golang.org/protobuf/proto"
)

const (
	MaxRequestBodySize = 10 * 1024 * 1024 // 10 MB
	IoBufferSize       = 1 * 1024 * 1024  // 1 MB
	IdleTimeout        = 30 * time.Minute
)

// Conn handles I/O operations for a network connection
type Conn struct {
	fd   int
	file *os.File
	conn net.Conn
}

// NewConn creates a new IOHandler from a file descriptor
func NewConn(clientFD int) (*Conn, error) {
	file := os.NewFile(uintptr(clientFD), fmt.Sprintf("client-fd-%d", clientFD))
	if file == nil {
		return nil, fmt.Errorf("failed to create file from file descriptor %d", clientFD)
	}

	var conn net.Conn
	defer func() {
		// Only close the file if we haven't successfully created a net.Conn
		if conn == nil {
			file.Close()
		}
	}()

	var err error
	conn, err = net.FileConn(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create net.Conn from file descriptor: %w", err)
	}

	if tcpConn, ok := conn.(*net.TCPConn); ok {
		if err := tcpConn.SetNoDelay(true); err != nil {
			return nil, fmt.Errorf("failed to set TCP_NODELAY: %w", err)
		}
		if err := tcpConn.SetKeepAlive(true); err != nil {
			return nil, fmt.Errorf("failed to set keepalive: %w", err)
		}
		if err := tcpConn.SetKeepAlivePeriod(time.Duration(90) * time.Second); err != nil {
			return nil, fmt.Errorf("failed to set keepalive period: %w", err)
		}
	}

	return &Conn{
		fd:   clientFD,
		file: file,
		conn: conn,
	}, nil
}

// ReadRequest reads data from the network connection
func (c *Conn) ReadSync() ([]byte, error) {
	reader := bufio.NewReaderSize(c.conn, IoBufferSize)

	// Step 1: Read 4 bytes for length prefix
	var lenBuf [4]byte
	if _, err := io.ReadFull(reader, lenBuf[:]); err != nil {
		return nil, err
	}
	length := binary.BigEndian.Uint32(lenBuf[:])

	// Validate length
	if length == 0 || length > MaxRequestBodySize {
		return nil, fmt.Errorf("invalid message length: %d", length)
	}

	// Step 2: Allocate buffer for the message
	msgBuf := make([]byte, length)

	// Step 3: Read the full message
	if _, err := io.ReadFull(reader, msgBuf); err != nil {
		return nil, err
	}

	return msgBuf, nil
}

func (c *Conn) WriteSync(r *comm.Response) error {

	// Encode protobuf
	data, err := proto.Marshal(r)
	if err != nil {
		return err
	}

	// Create a buffer
	var buf bytes.Buffer

	// Write the length prefix (4 bytes, big endian)
	if err := binary.Write(&buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}

	// Write the protobuf data
	if _, err := buf.Write(data); err != nil {
		return err
	}

	// Send everything
	_, err = c.conn.Write(buf.Bytes())
	return err
}

// Close underlying network connection
func (c *Conn) Close() error {
	var err error
	if c.conn != nil {
		err = errors.Join(err, c.conn.Close())
	}
	if c.file != nil {
		err = errors.Join(err, c.file.Close())
	}

	return err
}
