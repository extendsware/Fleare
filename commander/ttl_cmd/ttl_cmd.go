package ttl_cmd

import (
	"fmt"
	"strconv"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var ttlCmd = &common.Command{
	Name:        "TTL",
	Description: "Get the time to live for a key in seconds",
	Syntax:      "TTL <key>",
	Example: `
	localhost:9219> SETEX mykey 10 "Hello"
	OK
	localhost:9219> TTL mykey
	(integer) 10
	localhost:9219> TTL nonexistent
	(integer) -2
	`,
	Execute: ttlKey,
}

var expireCmd = &common.Command{
	Name:        "EXPIRE",
	Description: "Set a timeout on key. After the timeout has expired, the key will automatically be deleted",
	Syntax:      "EXPIRE <key> <seconds>",
	Example: `
	localhost:9219> SET mykey "Hello"
	OK
	localhost:9219> EXPIRE mykey 10
	(integer) 1
	localhost:9219> TTL mykey
	(integer) 10
	`,
	Execute: expireKey,
}

var setexCmd = &common.Command{
	Name:        "SETEX",
	Description: "Set key to hold the string value and set key to timeout after a given number of seconds",
	Syntax:      "SETEX <key> <seconds> <value>",
	Example: `
	localhost:9219> SETEX mykey 10 "Hello"
	OK
	localhost:9219> TTL mykey
	(integer) 10
	localhost:9219> GET mykey
	"Hello"
	`,
	Execute: setexKey,
}

func init() {
	common.Register(ttlCmd.Name, ttlCmd)
	common.Register(expireCmd.Name, expireCmd)
	common.Register(setexCmd.Name, setexCmd)
}

func ttlKey(cmd *common.Cmd) (*common.CmdResponse, error) {
	if cmd.C.Args == nil || len(cmd.C.Args) != 1 {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	ttl, err := shard.M.TTL(key)
	if err != nil {
		return nil, err
	}

	result := fmt.Sprintf("(integer) %d", ttl)
	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(result),
		},
	}, nil
}

func expireKey(cmd *common.Cmd) (*common.CmdResponse, error) {
	if cmd.C.Args == nil || len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: Key and seconds must be provided", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	secondsStr := cmd.C.Args[1]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: Invalid seconds value", errors.InvalidArgsError)
	}

	shard := cmd.SM.GetShardByKey(key)
	success, err := shard.M.Expire(key, seconds)
	if err != nil {
		return nil, err
	}

	var result string
	if success {
		result = "(integer) 1"
		// Log to WAL if expiration was set successfully
		obj, _ := shard.M.Get(key)
		if obj != nil {
			cmd.SM.Wal().Put(key, obj)
		}
	} else {
		result = "(integer) 0"
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(result),
		},
	}, nil
}

func setexKey(cmd *common.Cmd) (*common.CmdResponse, error) {
	if cmd.C.Args == nil || len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: Key, seconds, and value must be provided", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	secondsStr := cmd.C.Args[1]
	value := utils.StringToByte(cmd.C.Args[2])

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	seconds, err := strconv.ParseInt(secondsStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("%s: Invalid seconds value", errors.InvalidArgsError)
	}

	shard := cmd.SM.GetShardByKey(key)
	if err = shard.M.SetWithTTL(key, value, store.Default, seconds); err != nil {
		return nil, err
	}

	// Log to WAL
	obj, _ := shard.M.Get(key)
	if obj != nil {
		cmd.SM.Wal().Put(key, obj)
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte("OK"),
		},
	}, nil
}
