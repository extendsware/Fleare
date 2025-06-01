package eventloop

import (
	"context"

	"github.com/parashmaity/fleare/internal/logger"
)

type IOThread struct {
	ClientID string
	Mode     string
	IOConn   *Conn
	S        IOServer
}

func NewIOThread(clientFD int, ClientID string, sv IOServer) (*IOThread, error) {
	io, err := NewConn(clientFD)
	if err != nil {
		logger.Error("Failed to create new IOHandler for clientFD", err, map[string]any{"client-fd": clientFD})
		return nil, err
	}
	return &IOThread{
		ClientID: ClientID,
		Mode:     "sync",
		IOConn:   io,
		S:        sv,
	}, nil
}

func (t *IOThread) StartSync(ctx context.Context) error {
	for {
		c, err := t.IOConn.ReadSync()
		if err != nil {
			return err
		}

		t.S.OnTraffic(*t.IOConn, t.ClientID, c)
	}
}

func (t *IOThread) Stop() error {
	return t.IOConn.Close()
}
