package common

import (
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var setCmd = &Command{
	Name:        "SET",
	Description: "Set a key-value pair",
	Example: `
	localhost:9219> set key value
	Ok
	localhost:9219> get key
	value
	`,
	Execute: setKey,
}

func init() {
	Register(setCmd.Name, setCmd)
}

func setKey(cmd *Cmd) (*CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: Key and value must be provided", errors.InvalidArgsError)
	}
	key := cmd.C.Args[0]
	value := utils.StringToByte(cmd.C.Args[1])

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	if err = shard.M.Set(key, value, store.Default); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: value, Kind: uint32(store.Default)})

	return &CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
