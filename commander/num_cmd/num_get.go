package num_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var numberGetCmd = &common.Command{
	Name:        "NUM.GET",
	Description: "Get a number value from the database, The value return only number type (int, float, etc.) or error",
	Syntax:      "NUM.GET <key>",
	Example: `
	localhost:9219> NUM.SET myNumber 42
	Ok 
	localhost:9219> NUM.GET myNumber
	Ok 42
	localhost:9219> NUM.SET myNumber 11.5
	Ok 
	localhost:9219> GET myNumber
	Ok 11.5
	`,
	Execute: numberGetFunc,
}

func init() {
	common.Register(numberGetCmd.Name, numberGetCmd)
}

func numberGetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 1 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	if obj != nil {
		if store.Number != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: the existing value for the provided key must be a number", errors.InvalidValueError)
		}
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   obj.Value,
		},
	}, nil
}
