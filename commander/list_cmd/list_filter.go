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

// TODO
// Array filtering inside maps (e.g., $tags:contains:urgent)
// Regex match on fields (already partially done with ~=re)

var listFilterCmd = &common.Command{
	Name:        "LIST.FILTER",
	Description: "filter the list of element(s) by using path. An error is returned for out of range indexes.",
	Syntax:      "LIST.FILTER <key> <path>",
	Example: `
	127.0.0.1:9219> LIST.SET myArray "John" "Emily" "Michael" "Sarah" "David" "Jessica" "Robert" "Lisa" "James" "Jennifer" "John" "Emily" "Michael" "Sarah" "David" "Jessica" "Robert" "Lisa" "James" "Jennifer"
	Ok 20

	127.0.0.1:9219> LIST.FILTER myArray ':John'
	Ok ["John","John"]

	127.0.0.1:9219> LIST.FILTER myArray ':John'
	Ok ["Emily","Michael","Sarah","David","Jessica","Robert","Lisa","James","Jennifer","Emily","Michael","Sarah","David","Jessica","Robert","Lisa","James","Jennifer"]
	
	127.0.0.1:9219> LIST.FILTER myArray '~jessica'
	Ok ["Jessica","Jessica"]

	127.0.0.1:9219> LIST.SET myNum 10 20 30 40 50 60 70 30 20 40 20 70
	Ok 12

	127.0.0.1:9219> LIST.FILTER myNum ':20'
	Ok [20,20,20]

	127.0.0.1:9219> LIST.FILTER myNum '!20'
	Ok [10,30,40,50,60,70,30,40,70]

	127.0.0.1:9219> LIST.FILTER myNum '>50'
	Ok [60,70,70]

	127.0.0.1:9219> LIST.FILTER myNum '<=30'
	Ok [10,20,30,30,20,20]
	
	127.0.0.1:9219> LIST.PUSH myObj '{"name":"John", "age":30, "city":"New York"}'
	Ok 1

	127.0.0.1:9219> LIST.PUSH myObj '{"name":"Robert", "age":40, "city":"New York", "preferences": {"theme": "dark"}}'
	Ok 2

	127.0.0.1:9219> LIST.PUSH myObj '{"name":"David", "age":28, "city":"Kolkata", "preferences": {"theme": "dark"}}'
	Ok 3

	127.0.0.1:9219> LIST.PUSH myObj '{"name":"Michael", "age":25, "city":"Kolkata", "preferences": {"theme": "white"}}'
	Ok 4

	127.0.0.1:9219> LIST.FILTER myObj '$name:John'
	Ok [{"age":30,"city":"New York","name":"John"}]

	127.0.0.1:9219> LIST.FILTER myObj '$city:New York'
	Ok [{"age":30,"city":"New York","name":"John"},{"age":40,"city":"New York","name":"Robert","preferences":{"theme":"dark"}}]

	127.0.0.1:9219> LIST.FILTER myObj '$age>=28'
	Ok [{"age":30,"city":"New York","name":"John"},{"age":40,"city":"New York","name":"Robert","preferences":{"theme":"dark"}},{"age":28,"city":"Kolkata","name":"David","preferences":{"theme":"dark"}}]

	127.0.0.1:9219> LIST.FILTER myObj '$preferences.theme:white'
	Ok [{"age":25,"city":"Kolkata","name":"Michael","preferences":{"theme":"white"}}]

	127.0.0.1:9219> LIST.FILTER myObj '$preferences.theme!white'
	Ok [{"age":40,"city":"New York","name":"Robert","preferences":{"theme":"dark"}},{"age":28,"city":"Kolkata","name":"David","preferences":{"theme":"dark"}}]
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

func Filter(data []interface{}, path string) []interface{} {
	field, op, value := utils.ParsePath(path)
	var result []interface{}

	for _, item := range data {
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

// getNestedValue extracts value from nested maps using dot notation (e.g., "preferences.theme")
func getNestedValue(m map[string]interface{}, path string) (string, bool) {
	parts := strings.Split(path, ".")
	var val any = m
	for _, part := range parts {
		if mm, ok := val.(map[string]interface{}); ok {
			val = mm[part]
		} else {
			return "", false
		}
	}
	return fmt.Sprintf("%v", val), true
}

func match(left, op, right string) bool {
	switch op {
	case ":":
		return valueEqual(left, right)
	case "!":
		return !valueEqual(left, right)
	case "~":
		return strings.EqualFold(left, right)
	case ">", "<", ">=", "<=":
		return compareFloat(left, right, op)
	default:
		return false
	}
}

func valueEqual(a, b string) bool {
	// Try boolean
	if ab, err1 := strconv.ParseBool(a); err1 == nil {
		if bb, err2 := strconv.ParseBool(b); err2 == nil {
			return ab == bb
		}
	}

	// Try float
	if af, err1 := strconv.ParseFloat(a, 64); err1 == nil {
		if bf, err2 := strconv.ParseFloat(b, 64); err2 == nil {
			return af == bf
		}
	}

	// Default string compare
	return a == b
}

func compareFloat(aStr, bStr, op string) bool {
	a, err1 := strconv.ParseFloat(aStr, 64)
	b, err2 := strconv.ParseFloat(bStr, 64)
	if err1 != nil || err2 != nil {
		return false
	}

	switch op {
	case ">":
		return a > b
	case "<":
		return a < b
	case ">=":
		return a >= b
	case "<=":
		return a <= b
	default:
		return false
	}
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
