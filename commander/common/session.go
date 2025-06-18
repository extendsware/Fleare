package common

import (
	"github.com/parashmaity/fleare/internal/auth"
	"github.com/parashmaity/fleare/internal/comm"
)

var sessionCmd = &Command{
	Name:        "SESSION",
	Description: "Return current Session",
	Syntax:      "SESSION",
	Example: `
	localhost:9219> session
	Ok {
	  "created_at": "2025-05-01T17:33:15.497273Z",
	  "last_accessed_at": "2025-05-01T17:33:15.497273Z",
	  "session_id": "8-127.0.0.1:53531",
	  "status": 1,
	  "user": {
	    "Password": "*******",
	    "Role": "Admin",
	    "Username": "admin"
	  }
	}`,
	Execute: sessionAdd,
}

func init() {
	Register(sessionCmd.Name, sessionCmd)
}

func sessionAdd(cmd *Cmd) (*CmdResponse, error) {

	d, err := auth.SessionsStore.Store[cmd.ClientID].String()
	if err != nil {
		return nil, err
	}
	return &CmdResponse{
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(d),
		},
	}, nil
}
