package common

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

// TODO
// Array filtering inside maps (e.g., $tags:contains:urgent)
// Regex match on fields (already partially done with ~=re)

var listUpdateCmd = &common.Command{
	Name:        "LIST.UPDATE",
	Description: "Update the list of element(s) by using path. Return a positive number for updated keys count.",
	Syntax:      "LIST.UPDATE <key> <path> <value>",
	Example: `
	127.0.0.1:9219> LIST.SET myArray "John" "Emily" "Michael" "Sarah" "David" "Jessica" "Robert" "Lisa" "James" "Jennifer" "John" "Emily" "Michael" "Sarah" "David" "Jessica" "Robert" "Lisa" "James" "Jennifer"
	Ok 20

	127.0.0.1:9219> LIST.UPDATE myArray ':John' "Parash"
	Ok 2

	`,
	Execute: listUpdateFunc,
}

func init() {
	common.Register(listUpdateCmd.Name, listUpdateCmd)
}

func listUpdateFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: LIST.UPDATE <key> <path> <value>", errors.InvalidArgsError)
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
	var arr []any
	if obj == nil {
		return sentResponse([]byte(""), nil)
	}

	if store.List != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a list", errors.InvalidValueError)
	}
	if err := json.Unmarshal(obj.Value, &arr); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	if len(arr) == 0 {
		return sentResponse([]byte(""), nil)
	}

	n := FilterAndUpdate(arr, path, utils.EnsureUnmarshal(value))

	objBytes := utils.ObjectToByte(arr)

	obj = &comm.Object{Value: objBytes, Kind: uint32(store.List)}
	if err = shard.M.Set(key, objBytes, store.List); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, obj)

	return sentResponse([]byte(fmt.Sprint(n)), nil)
}

func FilterAndUpdate(data []interface{}, path string, newValue interface{}) int {
	field, op, value := utils.ParsePath(path)

	n := 0
	for i, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			if field == "" {
				continue
			}
			val, ok := getNestedValue(v, field)
			if !ok {
				continue
			}
			if match(val, op, value) {
				data[i] = newValue
				n++
			}
		default:
			if field != "" {
				continue
			}
			val := fmt.Sprintf("%v", v)
			if match(val, op, value) {
				data[i] = newValue
				n++
			}
		}
	}
	return n
}
