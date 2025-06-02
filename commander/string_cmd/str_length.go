package string_cmd

import (
	"fmt"
	"strconv"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var strLength = &common.Command{
	Name:        "STR.LENGTH",
	Description: "STR.LENGTH retrieve length of the string value from the key. If the key does not exist, it will be return 0.",
	Syntax:      "STR.LENGTH <key>",
	Example: `
	localhost:9219> STR.SET key1 "Hello"
	Ok
	localhost:9219> LENGTH key1 
	Ok 5
	localhost:9219> STR.LENGTH key2
	Ok 0
	`,
	Execute: strLengthFunc,
}

func init() {
	common.Register(strLength.Name, strLength)
}

func strLengthFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

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

	length := 0
	obj, _ := shard.M.Get(key)
	if obj != nil {
		if store.String != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: the existing value for the provided key must be a string", errors.InvalidValueError)
		}

		length = len(obj.Value)
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(strconv.Itoa(length)),
		},
	}, nil
}
