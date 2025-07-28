package common

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var listPopCmd = &common.Command{
	Name:        "LIST.POP",
	Description: `the command pops a single element from the beginning of the list of the list stored at the specified key`,

	Syntax: "LIST.POP <key>",
	Example: `
	localhost:9219> LIST.PUSH myKey "This is my first element"
	Ok
	localhost:9219> LIST.PUSH myKey '{"name":"John", "address": "kolkata"}' 10023.22
	Ok
	localhost:9219> LIST.POP myKey
	Ok 10023.22
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element"
		{
			"name":"John",
			"address": "kolkata"
		}
	]
	`,
	Execute: listPopFunc,
}

func init() {
	common.Register(listPopCmd.Name, listPopCmd)
}

func listPopFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	var arr []any
	if obj != nil {
		if err := json.Unmarshal(obj.Value, &arr); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}
		if store.List != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: the existing value for the provided key must be a list", errors.InvalidValueError)
		}
	} else {
		return sentResponse([]byte(""), nil)
	}

	length := len(arr)
	if length == 0 {
		return sentResponse([]byte(""), nil)
	}
	if length == 1 {
		shard.M.Delete(key)
		cmd.SM.Wal().Delete(key)

		return sentResponse(utils.ObjectToByte(arr[0]), nil)
	}
	last := arr[length-1]
	arr = arr[:length-1]
	if last == nil {
		return sentResponse([]byte(""), nil)
	}

	objBytes := utils.ObjectToByte(arr[:length-1])
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.List)})

	return sentResponse(utils.ObjectToByte(last), nil)
}
