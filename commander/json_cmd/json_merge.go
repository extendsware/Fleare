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

var jsonMergeCmd = &common.Command{
	Name: "JSON.MERGE",
	Description: `Merges the provided JSON object into the existing JSON object stored at the given key.
					If the key does not exist, it creates a new JSON object. Nested objects are merged recursively,
					and existing fields are overwritten by new values when there is a conflict.`,

	Syntax: "JSON.MERGE <key> <value>",
	Example: `
	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	localhost:9219> JSON.MERGE myObj '{"hobbies":["reading","hiking"]}'
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

	127.0.0.1:9219> JSON.MERGE myObj '{"name":"John Doe","age":30,"isActive":true,"salary":10000}'
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
	`,
	Execute: jsonMergeFunc,
}

func init() {
	common.Register(jsonMergeCmd.Name, jsonMergeCmd)
}

func jsonMergeFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: JSON.MERGE <key> <value>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	value := cmd.C.Args[1]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	obj, _ := shard.M.Get(key)
	if obj == nil {
		// If key doesn't exist, create new JSON object with the provided value
		unmarshaled := utils.EnsureUnmarshal(value)
		objBytes := utils.ObjectToByte(unmarshaled)
		if err = shard.M.Set(key, objBytes, store.JSON); err != nil {
			return nil, err
		}
		cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.JSON)})
		return sentResponse(objBytes, nil)
	}

	if store.JSON != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a json object", errors.InvalidValueError)
	}

	var existingObj map[string]interface{}
	if err := json.Unmarshal(obj.Value, &existingObj); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	// Parse the new value to merge
	newValue := utils.EnsureUnmarshal(value)
	newObj, ok := newValue.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("%s: The value to merge must be a JSON object", errors.InvalidValueError)
	}

	// Merge the objects
	mergedObj := mergeObjects(existingObj, newObj)

	objBytes := utils.ObjectToByte(mergedObj)
	if err = shard.M.Set(key, objBytes, store.JSON); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.JSON)})
	return sentResponse(objBytes, nil)
}

func mergeObjects(existing, new map[string]interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	// Copy existing object
	for k, v := range existing {
		result[k] = v
	}

	// Merge new object, recursively merging nested objects
	for k, v := range new {
		if existingVal, exists := result[k]; exists {
			// If both values are maps, merge them recursively
			if existingMap, ok := existingVal.(map[string]interface{}); ok {
				if newMap, ok := v.(map[string]interface{}); ok {
					result[k] = mergeObjects(existingMap, newMap)
					continue
				}
			}
		}
		// Otherwise, overwrite with new value
		result[k] = v
	}

	return result
}
