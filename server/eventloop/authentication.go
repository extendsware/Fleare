package eventloop

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"

	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

func (s *Server) ConnectAndAuthenticate(clientFD int, clientID string) bool {
	auth, err := s.AuthenticateClient(clientFD, clientID)
	if err != nil {
		eMsg := fmt.Sprintf("%s: %v", ErrAcceptConnection, err.Error())
		con := &comm.Response{
			Status: "Error",
			Result: []byte(eMsg),
		}
		s.WriteErr(clientFD, con)
		return false
	}

	con := &comm.Response{
		ClientId: clientID,
		Status:   "Ok",
		Result:   []byte("Connected"),
	}
	s.WriteErr(clientFD, con)
	return auth
}

func (s *Server) AuthenticateClient(clientFD int, clientID string) (bool, error) {
	var lenBuf [4]byte
	totalRead := 0

	// Step 1: Read exactly 4 bytes for length prefix
	for totalRead < 4 {
		n, err := unix.Read(clientFD, lenBuf[totalRead:])
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				continue // try again
			}
			return false, err
		}
		if n == 0 {
			return false, fmt.Errorf("empty auth request")
		}
		totalRead += n
	}

	length := binary.BigEndian.Uint32(lenBuf[:])
	if length == 0 || length > MaxRequestBodySize {
		return false, fmt.Errorf("invalid authentication message")
	}

	// Step 2: Allocate buffer for full protobuf message
	msgBuf := make([]byte, length)
	totalRead = 0

	// Step 3: Read the full message
	for totalRead < int(length) {
		n, err := unix.Read(clientFD, msgBuf[totalRead:])
		if err != nil {
			if errors.Is(err, unix.EAGAIN) || errors.Is(err, unix.EWOULDBLOCK) {
				continue // try again
			}
			return false, err
		}
		if n == 0 {
			return false, fmt.Errorf("empty auth request")
		}
		totalRead += n
	}

	// Step 4: Unmarshal Protobuf auth message
	var authReq comm.Command // your generated protobuf struct
	if err := proto.Unmarshal(msgBuf, &authReq); err != nil {
		return false, fmt.Errorf("invalid data format")
	}

	if len(authReq.Args) == 0 {
		return s.IOServer.RequestAuth("", "", clientID)
	}

	if len(authReq.Args) != 2 {
		return false, fmt.Errorf("invalid arguments")
	}
	return s.IOServer.RequestAuth(authReq.Args[0], authReq.Args[1], clientID)
}

func (s *Server) WriteErr(clientFD int, msg proto.Message) error {
	// Encode protobuf
	data, err := proto.Marshal(msg)
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
	unix.Write(clientFD, buf.Bytes())
	return err
}
