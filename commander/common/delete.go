package common

import (
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var deleteCmd = &Command{
	Name:        "DELETE",
	Description: "Deletes a key-value pair from the database",
	Example: `
	localhost:9219> DELETE key
	Ok
	localhost:9219> DELETE key
	Ok
	`,
	Execute: deleteKey,
}

func init() {
	Register(deleteCmd.Name, deleteCmd)
}

func deleteKey(cmd *Cmd) (*CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) != 1 {
		return nil, fmt.Errorf("%s: Key must be provided, example: DELETE key", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	status, err := shard.M.Delete(key)
	if err != nil {
		return nil, err
	}

	if status {
		cmd.SM.Wal().Delete(key)
	}

	return &CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
