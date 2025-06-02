package common

import (
	"fmt"
	"strings"

	"github.com/parashmaity/fleare/internal/comm"
)

var echoCmd = &Command{
	Name:        "PING",
	Description: "PONG back the input",
	Example: `
	localhost:9219> ping
	Ok PONG
	localhost:9219> PING
	Ok PONG
	`,
	Execute: echoBack,
}

func init() {
	Register(echoCmd.Name, echoCmd)
}

func echoBack(cmd *Cmd) (*CmdResponse, error) {

	s := fmt.Sprintf("%s %s", "PONG", strings.Join(cmd.C.Args, " "))

	return &CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(s),
		},
	}, nil
}
