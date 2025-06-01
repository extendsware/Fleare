package serve

import (
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var numberDecrCmd = &Command{
	Name:        "NUM.DECR",
	Description: "Decrements number value by a specified amount, or set it if it doesn't exist",
	Syntax:      "NUM.DECR <key> <value>",
	Example: `
	localhost:9219> NUM.SET myNumber 42
	Ok 
	localhost:9219> NUM.DECR myNumber
	Ok 42
	localhost:9219> GET myNumber
	Ok 41
	localhost:9219> NUM.DECR myNumber 10
	Ok 31
	localhost:9219> GET myNumber
	Ok 31
	localhost:9219> NUM.SET myNumber 11.5
	Ok 
	localhost:9219> NUM.DECR myNumber
	Ok 10.5
	localhost:9219>  NUM.DECR myNumber 0.5
	Ok 10
	`,
	Execute: numberDecrFunc,
}

func init() {
	Register(numberDecrCmd.Name, numberDecrCmd)
}

func numberDecrFunc(cmd *Cmd) (*CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) > 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}
	var v float64 = 1
	if len(cmd.C.Args) == 2 {
		v, err = utils.ConvertToNumber(cmd.C.Args[1])
		if err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "provided value is not a valid number")
		}
	}

	shard := cmd.SM.GetShardByKey(key)
	obj, _ := shard.M.Get(key)
	if obj != nil {
		current, err := utils.ConvertToNumber(string(obj.Value))
		if err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "existing value is not a valid number")
		}
		v = current - v
	}
	byteValue := utils.ObjectToByte(v)

	if err = shard.M.Set(key, byteValue); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: byteValue})

	return &CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   byteValue,
		},
	}, nil
}
