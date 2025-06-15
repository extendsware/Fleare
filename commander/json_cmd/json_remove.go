package json_cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var jsonRemoveCmd = &common.Command{
	Name: "JSON.REMOVE",
	Description: `Removes a specific path or the entire JSON value associated with a key. If only the key is provided,
					the entire JSON object is deleted. If a path is provided, only the specified field within the JSON object is removed.
					Supports nested path removal using dot notation.`,

	Syntax: "JSON.REMOVE <key> [<path>]",
	Example: `
	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	localhost:9219> JSON.REMOVE myObj "address"
	Ok {
		"age": 30,
		"isActive": true,
		"name": "John Doe"
	}

	127.0.0.1:9219> JSON.REMOVE myObj 'isActive'
	Ok {
		"age": 30,
		"name": "John Doe"
	}

	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	127.0.0.1:9219> JSON.REMOVE myObj 'address.city'
	Ok {
		"age": 30,
		"name": "John Doe",
		"isActive": true,
		"address": {
			"street": "123 Main St"
		}
	}
	`,
	Execute: jsonRemoveFunc,
}

func init() {
	common.Register(jsonRemoveCmd.Name, jsonRemoveCmd)
}

func jsonRemoveFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

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
	if obj == nil {
		return sentResponse([]byte(""), nil)
	}

	if store.JSON != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a json object", errors.InvalidValueError)
	}

	if len(cmd.C.Args) == 1 {
		shard.M.Delete(key)
		cmd.SM.Wal().Delete(key)
		return sentResponse([]byte(""), nil)
	}

	var M map[string]interface{}
	if err := json.Unmarshal(obj.Value, &M); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}
	path := cmd.C.Args[1]
	d, err := removeNestedKey(M, path)
	if err != nil {
		return sentResponse([]byte(""), err)
	}

	objBytes := utils.ObjectToByte(d)
	if err = shard.M.Set(key, objBytes, store.JSON); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.JSON)})
	return sentResponse(objBytes, nil)
}

func removeNestedKey(obj map[string]interface{}, path string) (map[string]interface{}, error) {
	keys := strings.Split(path, ".")

	// If obj is nil, return as is
	if obj == nil {
		return obj, nil
	}

	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			// Remove the key at the final level
			delete(current, key)
			return obj, nil
		}

		// Traverse to the next nested map
		val, exists := current[key]
		if !exists {
			// Path doesn't exist, nothing to remove
			return obj, nil
		}

		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid path: %s", path)
		}
		current = nextMap
	}
	return obj, nil
}
