package common

import (
	"encoding/json"
	"time"

	"github.com/parashmaity/fleare/internal/comm"
)

type Status struct {
	ServerStatus string         `json:"server_status"`
	ShardInfo    map[string]any `json:"shard_info"`
	// Shards       map[string]*shard.Shard `json:"shards"`
	UpTime time.Time `json:"up_time"`
}

var statusCmd = &Command{
	Name:        "STATUS",
	Description: "Get status of the server",
	Example: `
	localhost:9219> status
	Ok {
	  "server_status": "running",
	  "shard_info": {
	    "0": {
	      "ID": "0",
	      "host_address": "localhost:9291",
	      "key_length": 123,
	      "name": "Shard 0",
	      "total_size": "8 bytes"
	    },
	    "1": {
	      "ID": "1",
	      "host_address": "localhost:9291",
	      "key_length": 454,
	      "name": "Shard 1",
	      "total_size": "8 bytes"
	    },
	    "shard_count": 2
	  }
	`,
	Execute: serverStatus,
}

func init() {
	Register(statusCmd.Name, statusCmd)
}

func serverStatus(cmd *Cmd) (*CmdResponse, error) {

	status := Status{
		ServerStatus: "running",
		ShardInfo:    cmd.SM.GetAllShardInfo(),
		// Shards:       cmd.SM.GetAllShardMap(),
		UpTime: time.Now(),
	}

	jsonData, err := json.Marshal(status)
	if err != nil {
		return nil, err
	}

	return &CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(jsonData),
		},
	}, nil
}
