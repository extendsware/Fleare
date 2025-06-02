package common

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var getCmd = &Command{
	Name:        "GET",
	Description: "Get a value by key",
	Example: `
	localhost:9219> GET key value
	Ok
	localhost:9219> GET key
	value
	`,
	Execute: getKey,
}

func init() {
	Register(getCmd.Name, getCmd)
}

func getKey(cmd *Cmd) (*CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) > 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, err := shard.M.Get(key)
	if err != nil {
		return nil, err
	}
	if obj == nil {
		return &CmdResponse{
			D: &comm.Response{
				Result: []byte(""),
			},
		}, nil
	}

	if len(cmd.C.Args) == 2 {
		var M map[string]any
		path := cmd.C.Args[1]
		if err := json.Unmarshal(obj.Value, &M); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}
		d, _ := readNestedKey(M, path)

		v := utils.ObjectToByte(d)
		return &CmdResponse{
			D: &comm.Response{
				Result: v,
			},
		}, nil
	}

	return &CmdResponse{
		D: &comm.Response{
			Result: obj.Value,
		},
	}, nil
}

func readNestedKey(obj map[string]any, path string) (any, error) {
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
		nextMap, ok := val.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("invalid path: %s", path)
		}
		current = nextMap
	}
	return nil, fmt.Errorf("unexpected end of path")
}
