package num_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var numberCmd = &common.Command{
	Name:        "NUM.SET",
	Description: "Set a number value in the database, Value can be any number type (int, float, etc.)",
	Syntax:      "NUM.SET <key> <value>",
	Example: `
	localhost:9219> NUM.SET myNumber 42
	Ok 
	localhost:9219> GET myNumber
	Ok 42
	localhost:9219> NUM.SET myNumber 11.993400032
	Ok 
	localhost:9219> GET myNumber
	Ok 11.993400032
	`,
	Execute: numberSetFunc,
}

func init() {
	common.Register(numberCmd.Name, numberCmd)
}

func numberSetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	value := cmd.C.Args[1]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	v, err := utils.ConvertToNumber(value)

	if err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "provided value is not a valid number")
	}
	byteValue := utils.ObjectToByte(v)
	shard := cmd.SM.GetShardByKey(key)
	if err = shard.M.Set(key, byteValue, store.Number); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: byteValue, Kind: uint32(store.Number)})

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(""),
		},
	}, nil
}
