package json_cmd

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var jsonSetCmd = &common.Command{
	Name: "JSON.SET",
	Description: `Stores a JSON object against the specified key. If the key already exists, 
				 it will be overwritten. The value must be a valid JSON object. This command is useful for managing structured data such as user profiles,
				 configurations, or nested objects in a flexible, schema-less format.`,

	Syntax: "JSON.SET <key> <value>",
	Example: `
	localhost:9219> JSON.SET myObj '{"name":"John","age":30,"hobbies":["reading","hiking"]}'
	Ok

	localhost:9219> JSON.GET myObj
	Ok {
		"age": 30,
		"hobbies": [
			"reading",
			"hiking"
		],
		"name": "John"
	}

	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"hobbies":["reading","hiking"],"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	localhost:9219> JSON.GET myObj
	Ok {
		"address": {
			"city": "New York",
			"street": "123 Main St"
		},
		"age": 30,
		"hobbies": [
			"reading",
			"hiking"
		],
		"isActive": true,
		"name": "John Doe"
	}
	`,
	Execute: jsonSetFunc,
}

func init() {
	common.Register(jsonSetCmd.Name, jsonSetCmd)
}

func jsonSetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: LIST.FILTER <key> <path>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	value := utils.StringToByte(cmd.C.Args[1])

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	if !json.Valid(value) {
		return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "provided value not a valid json object")
	}

	shard := cmd.SM.GetShardByKey(key)

	if err = shard.M.Set(key, value, store.JSON); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: value, Kind: uint32(store.JSON)})

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
