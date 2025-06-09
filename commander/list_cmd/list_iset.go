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

var listISetCmd = &common.Command{
	Name:        "LIST.ISET",
	Description: "Set the element at index to the list. An error is returned for out of range indexes.",
	Syntax:      "LIST.ISET <key> <index> <value>",
	Example: `
	localhost:9219> LIST.ISET myKey 0 "This is my first element"
	Ok
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element"
	]
	localhost:9219> LIST.ISET myKey 1 "This is my second element"
	Ok
	localhost:9219> LIST.GET myKey 1
	Ok "This is my second element"
	`,
	Execute: listISetFunc,
}

func init() {
	common.Register(listISetCmd.Name, listISetCmd)
}

func listISetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments,Syntax: LIST.SET <key> <index> <value>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	index, ok := utils.ParseInt(cmd.C.Args[1])
	if !ok || index < 0 {
		return nil, fmt.Errorf("%s: Invalid index value. The index must be a number greater than or equal to 0", errors.InvalidArgsError)
	}
	element := cmd.C.Args[2]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	data := utils.EnsureUnmarshal(element)
	var arr []any
	if obj == nil {
		if index > 0 {
			return nil, fmt.Errorf("%s: index out of range", errors.InvalidIndexError)
		}
		arr = append(arr, data)
	} else {

		if err := json.Unmarshal(obj.Value, &arr); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}
		if index >= len(arr) {
			return nil, fmt.Errorf("%s: index out of range", errors.InvalidIndexError)
		}
		arr[index] = data
	}

	objBytes := utils.ObjectToByte(arr)

	obj = &comm.Object{Value: objBytes, Kind: uint32(store.List)}
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}
	cmd.SM.Wal().Put(key, obj)

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
