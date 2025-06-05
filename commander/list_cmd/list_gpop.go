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

var listGPopCmd = &common.Command{
	Name:        "LIST.GPOP",
	Description: `(Get and Pop)Removes and returns the first element of the list stored at the specified key`,

	Syntax: "LIST.GPOP <key>",
	Example: `
	localhost:9219> LIST.PUSH myKey "This is my first element"
	Ok
	localhost:9219> LIST.PUSH myKey '{"name":"John", "address": "kolkata"}' 10023.22
	Ok
	localhost:9219> LIST.GPOP myKey
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
	Execute: listGPopFunc,
}

func init() {
	common.Register(listGPopCmd.Name, listGPopCmd)
}

func listGPopFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

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
	}
	last := arr[len(arr)-1]

	objBytes := utils.ObjectToByte(arr[:len(arr)-1])
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}
	obj.Value = objBytes
	cmd.SM.Wal().Put(key, obj)

	item := utils.ObjectToByte(last)

	return &common.CmdResponse{
		D: &comm.Response{
			Result: item,
		},
	}, nil
}
