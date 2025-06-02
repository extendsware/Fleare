package string_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var strGet = &common.Command{
	Name:        "STR.GET",
	Description: "STR.GET retrieve the string data value from the key. If the key does not exist, it will be return empty value.",
	Syntax:      "STR.GET <key>",
	Example: `
	localhost:9219> STR.SET key1 "Hello"
	Ok
	localhost:9219> GET key1 
	Ok "Hello"
	localhost:9219> STR.GET key2
	Ok ""
	`,
	Execute: strGetFunc,
}

func init() {
	common.Register(strGet.Name, strGet)
}

func strGetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 1 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	if obj != nil {
		if store.String != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: the existing value for the provided key must be a string", errors.InvalidValueError)
		}
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(obj.Value),
		},
	}, nil
}
