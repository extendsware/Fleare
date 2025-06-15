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

var jsonGetCmd = &common.Command{
	Name: "JSON.GET",
	Description: `Fetches a JSON value associated with a given key. Optionally, you can specify a JSON path to retrieve a nested value inside the JSON object. 
				  If no path is specified, the entire JSON object is returned.

				  - If the key does not exist, an empty response is returned.
				  - If the value associated with the key is not a valid JSON object, an error is returned.
				  - If the provided path does not exist or is invalid, an error is returned.`,

	Syntax: "JSON.GET <key> [<path>]",
	Example: `
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

	localhost:9219> JSON.GET myObj address
	Ok {
		"city": "New York",
		"street": "123 Main St"
	}

	localhost:9219> JSON.GET myObj address.city
	Ok "New York"

	localhost:9219> JSON.GET myObj hobbies
	Ok ["reading","hiking"]
	`,
	Execute: jsonGetFunc,
}

func init() {
	common.Register(jsonGetCmd.Name, jsonGetCmd)
}

func jsonGetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) > 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: JSON.GET <key> [<path>]", errors.InvalidArgsError)
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
		return sentResponse(obj.Value, nil)
	}

	var M map[string]interface{}
	path := cmd.C.Args[1]
	if err := json.Unmarshal(obj.Value, &M); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}
	d, _ := readNestedKey(M, path)

	v := utils.ObjectToByte(d)
	return sentResponse(v, nil)
}

func readNestedKey(obj map[string]interface{}, path string) (interface{}, error) {
	keys := strings.Split(path, ".")

	current := obj
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return nil, fmt.Errorf("key not found: %s", key)
		}

		if i == len(keys)-1 {
			// Return pointer to the final value so it can be modified
			return val, nil
		}

		// Expecting a nested map for further traversal
		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid path: %s", path)
		}
		current = nextMap
	}
	return nil, fmt.Errorf("unexpected end of path")
}

func sentResponse(data []byte, err error) (*common.CmdResponse, error) {
	if err != nil {
		return nil, err
	}
	return &common.CmdResponse{
		D: &comm.Response{
			Result: data,
		},
	}, err
}
