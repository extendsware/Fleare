package server

import (
	"context"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/config"
	"github.com/parashmaity/fleare/internal/auth"
	"github.com/parashmaity/fleare/internal/comm"
	iomanager "github.com/parashmaity/fleare/internal/io-manager"
	"github.com/parashmaity/fleare/internal/logger"
	"github.com/parashmaity/fleare/internal/shard"
	"github.com/parashmaity/fleare/internal/utils"
	"github.com/parashmaity/fleare/server/eventloop"

	"os"
	"os/signal"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sys/unix"
	"google.golang.org/protobuf/proto"
)

type EventManager struct {
	eventloop.BuildInIOServer
	EnableAuth   bool
	Config       *config.Configuration
	ShardManager *shard.ShardManager
}

func Start() {
	// Create a channel to receive OS signals
	sigs := make(chan os.Signal, 1)

	// Register for SIGINT (Ctrl+C), SIGTERM, and SIGHUP
	signal.Notify(sigs, unix.SIGINT, unix.SIGTERM, unix.SIGHUP)

	tm := iomanager.NewThreadManager()

	config := config.GetConfig()

	utils.PrintName2()

	// Wait for interrupt signal
	go func() {
		sig := <-sigs
		logger.Console("Received signal, initiating shutdown", zerolog.WarnLevel, map[string]any{"signal": sig.String()})
		tm.CancelAll()
	}()

	opts := &eventloop.Options{
		MaxConnections: config.Misc.MaxConnections,
		EnableAuth:     config.Security.EnableAuth,
	}
	s := &EventManager{
		Config:       config,
		ShardManager: shard.NewShardManager(config.Shard.ShardCount),
	}

	if config.Persistence.Enable {
		tm.Go("shard-recovery", func(ctx context.Context) {
			s.ShardManager.RecoverMemory(config.Persistence.Path)
		})
	}

	eventloop.Start(s, tm, opts)

	tm.Wait()

	// Stop accepting signals and clean up
	signal.Stop(sigs)
	close(sigs)

	// Final log message
	logger.Console("Server shutdown complete", zerolog.InfoLevel, nil)
}

func (m *EventManager) RequestAuth(username string, password string, clientID string) (bool, error) {
	logger.Debug("Request for Authentication", map[string]any{
		"enableAuth": m.Config.Security.EnableAuth,
		"username":   username,
		"password":   password,
	})

	if !m.Config.Security.EnableAuth {
		if username != "" || password != "" {
			return false, fmt.Errorf("it looks like you entered a username or password but the server isn’t using authentication right now")
		}
		user, err := auth.UserStore.ValidateDefault("root")
		if err != nil {
			return false, err
		}
		s := auth.NewSession(user, clientID)
		return s.Activate()
	}

	if username == "" || password == "" {
		return false, fmt.Errorf("failed to authenticate, please ensure your username and password are provided")
	}

	user, err := auth.UserStore.Validate(username, password)
	if err != nil {
		return false, err
	}

	s := auth.NewSession(user, clientID)
	return s.Activate()
}

func (m *EventManager) OnStart(ip string, port int) {

	logger.Debug("Server starting", map[string]any{
		"IP-Address:": ip,
		"port":        port,
	})
}

func (m *EventManager) OnConnection(conn eventloop.Conn, clientID string) {
	logger.Info("New connection", map[string]any{
		"clientID:": clientID,
	})

}

func (m *EventManager) OnDisconnect(clientID string) {
	logger.Info("Disconnected client", map[string]any{
		"clientID:": clientID,
	})
	auth.ExpireById(clientID)
}

func (m *EventManager) OnTraffic(conn eventloop.Conn, clientID string, data []byte) {

	start := time.Now()
	c := &comm.Command{}
	if err := proto.Unmarshal(data, c); err != nil {

		errMsg := fmt.Sprintf("invalid command format: %v\n", err)
		res := &comm.Response{
			Status:   config.STATUS_ERROR,
			ClientId: clientID,
			Result:   []byte(errMsg),
		}
		// sent error response if protocol error
		conn.WriteSync(res)
		logger.Error(errMsg, err, map[string]any{
			"clientID:":    clientID,
			"receivedSize": len(data),
			"status":       config.STATUS_ERROR,
			"duration":     time.Since(start).String(),
		})
		return
	}

	cmd := &common.Cmd{
		SM:       m.ShardManager,
		ClientID: clientID,
		C:        c,
	}

	res, err := cmd.Execute()
	if err != nil {
		errMsg := fmt.Sprintf("%v\n", err)
		res := &comm.Response{
			Status:   config.STATUS_ERROR,
			ClientId: clientID,
			Result:   []byte(errMsg),
		}
		// sent error response if execution error
		conn.WriteSync(res)
		logger.Error(errMsg, err, map[string]any{
			"clientID:":    clientID,
			"receivedSize": len(data),
			"status":       config.STATUS_ERROR,
			"duration":     time.Since(start).String(),
		})
		return
	}
	res.D.Status = config.STATUS_SUCCESS
	res.ClientID = cmd.ClientID
	// sent success response
	conn.WriteSync(res.D)

	logger.Info("Response", map[string]any{
		"command:":     cmd.C.Command,
		"clientID:":    clientID,
		"receivedSize": fmt.Sprintf("%d bytes", len(data)),
		"size":         fmt.Sprintf("%d bytes", len(res.D.Result)),
		"status":       res.D.Status,
		"duration":     time.Since(start).String(),
	})
}

func (m *EventManager) OnStop() {
	// fmt.Println("Server stopping.")
}
