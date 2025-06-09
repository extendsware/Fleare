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

var listLenCmd = &common.Command{
	Name:        "LIST.LEN",
	Description: "get the length of list of list. An error is returned for the key is not a list.",
	Syntax:      "LIST.LEN <key>",
	Example: `
	localhost:9219> LIST.ISET myKey 0 "This is my first element"
	Ok
	localhost:9219> LIST.ISET myKey 1 "{"name":"John", "address": "kolkata"}"
	Ok
	localhost:9219> LIST.LEN myKey
	Ok 2
	`,
	Execute: listLenFunc,
}

func init() {
	common.Register(listLenCmd.Name, listLenCmd)
}

func listLenFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 1 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: LIST.LEN <key>", errors.InvalidArgsError)
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
				Result: []byte("0"),
			},
		}, nil
	}

	if store.List != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a list", errors.InvalidValueError)
	}
	if err := json.Unmarshal(obj.Value, &arr); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	length := len(arr)
	objBytes := utils.ObjectToByte(length)

	return &common.CmdResponse{
		D: &comm.Response{
			Result: objBytes,
		},
	}, nil
}
