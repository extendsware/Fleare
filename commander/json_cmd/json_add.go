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

var jsonAddCmd = &common.Command{
	Name: "JSON.ADD",
	Description: `Adds a new key-value pair to an existing JSON object stored at the given key. 
				 If the JSON object does not exist, a new one will be created. The path can be nested using dot (.) notation 
				 to insert data at the desired depth. If any intermediate path does not exist, it will be created automatically. 
				 The value must be a valid JSON string or primitive (e.g., string, number, boolean, array).`,

	Syntax: "JSON.ADD <key> <path> <value>",
	Example: `
	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	localhost:9219> JSON.ADD myObj "hobbies" '["reading","hiking"]'
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

	127.0.0.1:9219> JSON.ADD myObj 'salary' 10000
	Ok {
		"address": {
			"city": "New York",
			"street": "123 Main St"
		},
		"age": 30,
		"hobbies": [
			"reading",
			"hiking",
			"cricket"
		],
		"isActive": true,
		"name": "John Doe",
		"salary": 10000
	}

	127.0.0.1:9219> JSON.ADD myObj 'office.members' 134

	Ok {
		"address": {
			"city": "New York",
			"street": "123 Main St"
		},
		"age": 30,
		"hobbies": [
			"reading",
			"hiking",
			"cricket"
		],
		"isActive": true,
		"name": "John Doe",
		"office": {
			"members": 134
		},
		"salary": 10000
	}
	`,
	Execute: jsonAddFunc,
}

func init() {
	common.Register(jsonAddCmd.Name, jsonAddCmd)
}

func jsonAddFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: JSON.ADD <key> <path> <value>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	path := cmd.C.Args[1]
	value := cmd.C.Args[2]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	obj, _ := shard.M.Get(key)
	if obj == nil {
		var M map[string]interface{}
		d, err := addNestedKey(M, path, utils.EnsureUnmarshal(value))
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

	if store.JSON != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a json object", errors.InvalidValueError)
	}

	var M map[string]interface{}
	if err := json.Unmarshal(obj.Value, &M); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}
	d, err := addNestedKey(M, path, utils.EnsureUnmarshal(value))
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

func addNestedKey(obj map[string]interface{}, path string, value interface{}) (map[string]interface{}, error) {
	keys := strings.Split(path, ".")

	// If obj is nil, initialize it
	if obj == nil {
		obj = make(map[string]interface{})
	}

	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			// Set the value at the final key
			current[key] = value
			return obj, nil
		}

		// Traverse or create nested map
		val, exists := current[key]
		if !exists {
			nextMap := make(map[string]interface{})
			current[key] = nextMap
			current = nextMap
			continue
		}

		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid path: %s", path)
		}
		current = nextMap
	}
	return obj, nil
}
