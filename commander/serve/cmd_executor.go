package serve

import (
	"fmt"
	"strings"

	"github.com/parashmaity/fleare/internal/comm"
	comm_errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/shard"
)

// ExecCommands is a map of executable commands.
var ExecCommands = map[string]*Command{}

type CmdResponse struct {
	ClientID string
	ReqID    string
	D        *comm.Response
}

type Cmd struct {
	SM       *shard.ShardManager
	ClientID string
	C        *comm.Command
}

type Command struct {
	Name        string
	Description string
	Syntax      string
	Args        []string
	Execute     func(com *Cmd) (*CmdResponse, error)
	Example     string
	SubCommands map[string]*Command
}

// function for add command to ExecCommands
func Register(name string, cmd *Command) {
	ExecCommands[name] = cmd
}

// function to execute command
func (c *Cmd) Execute() (*CmdResponse, error) {
	_c, ok := ExecCommands[strings.ToUpper(c.C.Command)]
	if !ok {
		return nil, fmt.Errorf("%v -> \"%s\"",
			comm_errors.MsgInvalidCommandError, c.C.Command)
	}
	res, err := _c.Execute(c)
	return res, err
}
