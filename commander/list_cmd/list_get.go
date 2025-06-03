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

var listGetCmd = &common.Command{
	Name:        "LIST.GET",
	Description: "get the list of element(s) at index to element. An error is returned for out of range indexes.",
	Syntax:      "LIST.GET <key> <index>",
	Example: `
	localhost:9219> LIST.SET myKey 0 "This is my first element"
	Ok
	localhost:9219> LIST.SET myKey 1 "{"name":"John", "address": "kolkata"}"
	Ok
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element",
		{"name":"John", "address": "kolkata"}
	]
	localhost:9219> LIST.GET myKey 1
	Ok
	localhost:9219> LIST.GET myKey 1
	Ok {
		"name":"John", 
		"address": "kolkata"
	}
	`,
	Execute: listGetFunc,
}

func init() {
	common.Register(listGetCmd.Name, listGetCmd)
}

func listGetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

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

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	var arr []any
	if obj == nil {
		return &common.CmdResponse{
			D: &comm.Response{
				Result: []byte(""),
			},
		}, nil
	}

	if store.List != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a list", errors.InvalidValueError)
	}
	if err := json.Unmarshal(obj.Value, &arr); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	if len(cmd.C.Args) == 2 {
		index, ok := utils.ParseInt(cmd.C.Args[1])
		if !ok || index < 0 {
			return nil, fmt.Errorf("%s: Invalid index value. The index must be a number greater than or equal to 0", errors.InvalidArgsError)
		}
		if index >= len(arr) {
			return nil, fmt.Errorf("%s: index out of range", errors.InvalidIndexError)
		}

		objBytes := utils.ObjectToByte(arr[index])

		return &common.CmdResponse{
			D: &comm.Response{
				Result: objBytes,
			},
		}, nil
	}

	objBytes := utils.ObjectToByte(arr)

	return &common.CmdResponse{
		D: &comm.Response{
			Result: objBytes,
		},
	}, nil
}
