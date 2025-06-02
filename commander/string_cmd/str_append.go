package string_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var strAppend = &common.Command{
	Name:        "STR.APPEND",
	Description: "Append the string data items to the end of the existing string value.",
	Syntax:      "STR.APPEND <key> <value> [<value> ...]",
	Example: `
	localhost:9219> STR.APPEND key1 "Hello"
	Ok
	localhost:9219> GET key1 
	Ok "Hello"
	localhost:9219> STR.APPEND key1 " World" " John"
	Ok 
	localhost:9219> GET key1 
	Ok "Hello World John"
	`,
	Execute: strAppendFunc,
}

func init() {
	common.Register(strAppend.Name, strAppend)
}

func strAppendFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	var str string = ""
	obj, _ := shard.M.Get(key)
	if obj != nil {
		if store.String != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: the existing value for the provided key must be a string", errors.InvalidValueError)
		}
		str = string(obj.Value)
	}
	for _, arg := range cmd.C.Args[1:] {
		str += arg
	}

	if err = shard.M.Set(key, []byte(str), store.String); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: []byte(str), Kind: uint32(store.String)})

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(""),
		},
	}, nil
}
