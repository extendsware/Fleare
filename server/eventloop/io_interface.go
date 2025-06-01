package eventloop

type Options struct {
	// Add any additional options here
	EnableAuth bool

	// MaxConnections is the maximum number of connections allowed
	MaxConnections int

	// MaxBufferSize is the maximum size of a buffer
	MaxBufferSize int
}

type IOServer interface {
	// OnStart starts the server on the given address and port
	OnStart(ip string, port int)

	// OnConnection is called when a new connection is established
	RequestAuth(username string, password string, clientID string) (bool, error)

	OnConnection(conn Conn, clientID string)

	// OnDisconnect is called when a connection is closed
	OnDisconnect(clientID string)

	// OnData is called when data is received from a connection
	OnTraffic(conn Conn, clientID string, data []byte)

	// OnStop stops the server and closes all connections
	OnStop()
}

// BuildInIOServer provides a default implementation of the IOServer interface
type BuildInIOServer struct{}

// OnStart implements the IOServer.OnStart method
func (s *BuildInIOServer) OnStart(ip string, port int) {}

// OnConnection implements the IOServer.OnConnection method
func (s *BuildInIOServer) OnConnection(conn Conn, clientID string) {}

// OnConnection implements the IOServer.OnConnection method
func (s *BuildInIOServer) RequestAuth(username string, password string, clientID string) (bool, error) {
	return false, nil
}

// OnDisconnect implements the IOServer.OnDisconnect method
func (s *BuildInIOServer) OnDisconnect(clientID string) {}

// OnTraffic implements the IOServer.OnTraffic method
func (s *BuildInIOServer) OnTraffic(conn Conn, clientID string, data []byte) {}

// OnStop implements the IOServer.OnStop method
func (s *BuildInIOServer) OnStop() {}
