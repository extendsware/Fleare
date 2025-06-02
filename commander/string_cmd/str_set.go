package string_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var strSet = &common.Command{
	Name:        "STR.SET",
	Description: "STR.SET store the string data item in the key. If the key already exists, it will be overwritten with the new value.",
	Syntax:      "STR.SET <key> <value>",
	Example: `
	localhost:9219> STR.SET key1 "Hello"
	Ok
	localhost:9219> GET key1 
	Ok "Hello"
	localhost:9219> STR.SET key1 "World World"
	Ok 
	localhost:9219> GET key1 
	Ok "Hello World"
	`,
	Execute: strSetFunc,
}

func init() {
	common.Register(strSet.Name, strSet)
}

func strSetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	value := utils.StringToByte(cmd.C.Args[1])

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	if err = shard.M.Set(key, value, store.String); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: value, Kind: uint32(store.String)})

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(""),
		},
	}, nil
}
