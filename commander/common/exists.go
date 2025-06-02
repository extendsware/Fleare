package common

import (
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
)

var existsCmd = &Command{
	Name:        "EXISTS",
	Description: "Check if a key exists in the database",
	Syntax:      "EXISTS <key> [<key> ...]",
	Example: `
	localhost:9219> EXISTS key1
	Ok 1
	localhost:9219> EXISTS key1 key2
	Ok 2
	localhost:9219> EXISTS key1 key2 key3
	Ok 2
	`,
	Execute: existsFunc,
}

func init() {
	Register(existsCmd.Name, existsCmd)
}

func existsFunc(cmd *Cmd) (*CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	var found int8 = 0
	for _, key := range cmd.C.Args {

		shard := cmd.SM.GetShardByKey(key)
		obj, _ := shard.M.Get(key)
		if obj != nil {
			found++
		}
	}

	return &CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(fmt.Sprintf("%d", found)),
		},
	}, nil
}
