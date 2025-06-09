package common

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var listFilterCmd = &common.Command{
	Name:        "LIST.FILTER",
	Description: "filter the list of element(s) by using path. An error is returned for out of range indexes.",
	Syntax:      "LIST.FILTER <key> <path>",
	Example: `
	localhost:9219> LIST.PUSH myKey "One" "Two" "Three" "Four" 
	Ok
	
	localhost:9219> LIST.PUSH myKey '{"name":"John", "address": "kolkata"}'
	Ok

	localhost:9219> LIST.FILTER myKey $Two
	Ok [Two]

	localhost:9219> LIST.FILTER myKey $address:kolkata
	Ok [{
		"name":"John", 
		"address": "kolkata"
	}]
	`,
	Execute: listFilterFunc,
}

func init() {
	common.Register(listFilterCmd.Name, listFilterCmd)
}

func listFilterFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: LIST.FILTER <key> <path>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	path := cmd.C.Args[1]

	q, _ := utils.ParseQuery(path)
	fmt.Println(q.ArrayFilter, q.Field, q.Value)

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

	f := Filter(arr, path)

	objBytes := utils.ObjectToByte(f)

	return sentResponse(objBytes, nil)
}

// Filter filters data based on simplified path expressions
func Filter(data []interface{}, path string) []interface{} {
	field, op, value := utils.ParsePath(path)
	fmt.Println(field, op, value)
	var result []interface{}

	for _, item := range data {
		switch v := item.(type) {
		case map[string]interface{}:
			if field == "" {
				continue
			}
			val := fmt.Sprintf("%v", v[field])
			if match(val, op, value) {
				result = append(result, item)
			}
		default:
			if field != "" {
				continue
			}
			val := fmt.Sprintf("%v", v)
			if match(val, op, value) {
				result = append(result, item)
			}
		}
	}
	return result
}

// match compares two values using an operator
func match(left, op, right string) bool {
	switch op {
	case ":":
		return left == right
	case "!":
		return left != right
	case "~":
		return strings.EqualFold(left, right)
	case ">":
		return compareNum(left, right, func(l, r int) bool { return l > r })
	case "<":
		return compareNum(left, right, func(l, r int) bool { return l < r })
	case ">=":
		return compareNum(left, right, func(l, r int) bool { return l >= r })
	case "<=":
		return compareNum(left, right, func(l, r int) bool { return l <= r })
	default:
		return false
	}
}

// compareNum compares two integers with a custom function
func compareNum(lstr, rstr string, fn func(int, int) bool) bool {
	l, err1 := strconv.Atoi(lstr)
	r, err2 := strconv.Atoi(rstr)
	if err1 != nil || err2 != nil {
		return false
	}
	return fn(l, r)
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
