package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"test/comm"
	"time"

	"google.golang.org/protobuf/proto"
)

func main() {
	serverAddr := "127.0.0.1:9219" // change port if needed

	conn, err := net.Dial("tcp", serverAddr)
	if err != nil {
		fmt.Println("Failed to connect:", err)
		os.Exit(1)
	}
	defer conn.Close()

	// Build the Command message
	cmd := &comm.Command{
		Command: "echo",
		Args:    []string{"test-message"},
	}

	start := time.Now()
	WriteMessage(conn, cmd)

	var resp comm.Response
	err = ReadMessage(conn, &resp)

	fmt.Printf("%d %s %s\n", time.Since(start).Microseconds(), resp.Status, string(resp.Result))

	// // Print the response
	// fmt.Println("Response received, time:", time.Since(start))
	// fmt.Println("Client ID:", resp.ClientId)
	// fmt.Println("Status:", resp.Status)
	// fmt.Println("Result:", string(resp.Result))
}

// Write a framed Protobuf message
func WriteMessage(conn net.Conn, msg proto.Message) error {
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

	// Send everything
	_, err = conn.Write(buf.Bytes())
	return err
}

// Read a framed Protobuf message
func ReadMessage(conn net.Conn, msg proto.Message) error {
	var length uint32

	// Read the length prefix
	if err := binary.Read(conn, binary.BigEndian, &length); err != nil {
		return err
	}

	// Read the message data
	data := make([]byte, length)
	_, err := conn.Read(data)
	if err != nil {
		return err
	}

	// Unmarshal the data
	return proto.Unmarshal(data, msg)
}
