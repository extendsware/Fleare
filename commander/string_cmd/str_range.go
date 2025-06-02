package string_cmd

import (
	"fmt"
	"unicode/utf8"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var strRangeCmd = &common.Command{
	Name: "RANGE",
	Description: `Returns a substring of the string value stored at the specified key, based on the provided start and end offsets (both inclusive).
				  Negative offsets are supported and refer to positions counted from the end of the string. For example, -1 represents the last character, -2 the second-to-last, and so on.`,
	Example: `
	Syntax: RANGE key start end
	Example:
	localhost:9219> SET myKey "There are many variations of passages"
	Ok
	localhost:9219> RANGE myKey 0 5
	Ok "There"
	localhost:9219> RANGE myKey 15 10
	Ok "variations"
	localhost:9219> RANGE myKey -8 4
	Ok "pass"
	localhost:9219> RANGE myKey 15 -1
	Ok "variations of passages"
	`,
	Execute: strRangeFunc,
}

func init() {
	common.Register(strRangeCmd.Name, strRangeCmd)
}

func strRangeFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	start, ok := utils.ParseInt(cmd.C.Args[1])
	if !ok {
		return nil, fmt.Errorf("%s: %s", errors.InvalidArgsError, "start offsets must be valid integers")
	}
	end, ok := utils.ParseInt(cmd.C.Args[2])
	if !ok {
		return nil, fmt.Errorf("%s: %s", errors.InvalidArgsError, "end offsets must be valid integers")
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)
	if obj == nil {
		return nil, fmt.Errorf("%s: %s", errors.KeyNotFoundError, "provided key does not exist")
	}

	if !utf8.Valid(obj.Value) {
		return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "stored value is not a string for the provided key")
	}

	str := getRange(utils.ByteToString(obj.Value), start, end)

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(str),
		},
	}, nil
}

func getRange(value string, start, end int) string {

	runes := []rune(value)
	length := len(runes)

	// Normalize negative indices
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length
	} else {
		end = start + end
	}

	// Clamp indices
	if start < 0 {
		start = 0
	}

	if end >= length {
		end = length
	}
	if start > length {
		start = length
	}

	println(length, start, end)

	return string(runes[start:end])
}
