package common

import (
	"fmt"
	"strconv"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var listSetCmd = &common.Command{
	Name:        "LIST.SET",
	Description: "Clear all existing elements and Set the new element to the list. Returned indexes number.",
	Syntax:      "LIST.SET <key> [<value>]",
	Example: `
	localhost:9219> LIST.SET myKey "This is my first element"
	Ok 0
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element"
	]
	localhost:9219> LIST.SET myKey "This is my second element"
	Ok 0
	localhost:9219> LIST.GET myKey
	Ok [
		"This is my first element"
	]
	`,
	Execute: listSetFunc,
}

func init() {
	common.Register(listSetCmd.Name, listSetCmd)
}

func listSetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) < 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments,Syntax: LIST.SET <key> [<value>]", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	var arr []any
	for _, arg := range cmd.C.Args[1:] {
		data := utils.EnsureUnmarshal(arg)
		arr = append(arr, data)
	}

	objBytes := utils.ObjectToByte(arr)

	obj := &comm.Object{Value: objBytes, Kind: uint32(store.List)}
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}
	cmd.SM.Wal().Put(key, obj)

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(strconv.Itoa(len(arr))),
		},
	}, nil
}
