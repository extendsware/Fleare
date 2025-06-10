package common

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var listPushCmd = &common.Command{
	Name: "LIST.PUSH",
	Description: `Insert all the given values at the beginning of the list stored at the specified key.
				 If the key does not exist, an empty list is created before performing the insertion. 
				 Returns an error if the key exists but does not hold a list.`,

	Syntax: "LIST.PUSH <key> <element> [<element>...]",
	Example: `
	localhost:9219> LIST.PUSH myKey "This is my first element"
	Ok
	localhost:9219> LIST.PUSH myKey '{"name":"John", "address": "kolkata"}' 10023.22
	Ok
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element",
		{
		   "name":"John",
		   "address": "kolkata"
		},
		10023.22
	]
	`,
	Execute: listPushFunc,
}

func init() {
	common.Register(listPushCmd.Name, listPushCmd)
}

func listPushFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

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
	}
	for _, arg := range cmd.C.Args[1:] {
		data := utils.EnsureUnmarshal(arg)
		arr = append(arr, data)
	}

	objBytes := utils.ObjectToByte(arr)
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}
	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.List)})

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(strconv.Itoa(len(arr))),
		},
	}, nil
}
