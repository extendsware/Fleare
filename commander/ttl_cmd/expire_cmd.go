package ttl_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var expireCmd = &common.Command{
	Name: "TTL.EXPIRE",
	Description: `Sets a time-to-live (TTL) on a key, after which the key will expire and be automatically deleted.
					The duration can be specified in seconds (s), minutes (m), hours (h), or days (d).
					Returns 1 if the TTL was set successfully, 0 if the key does not exist.`,

	Syntax: "TTL.EXPIRE <key> <seconds[s]|minutes[m]|hours[h]|days[d]>",
	Example: `
	localhost:9219> SET mykey "Hello"
	OK
	localhost:9219> TTL.EXPIRE mykey 10
	Ok 1
	localhost:9219> TTL.EXPIRE mykey 10s
	Ok 1
	localhost:9219> TTL.EXPIRE mykey 10m
	Ok 1
	localhost:9219> TTL.EXPIRE mykey 2h
	Ok 1
	localhost:9219> TTL.EXPIRE mykey 1d
	Ok 1
	`,
	Execute: expireKey,
}

func init() {
	common.Register(expireCmd.Name, expireCmd)
}

func expireKey(cmd *common.Cmd) (*common.CmdResponse, error) {
	if cmd.C.Args == nil || len(cmd.C.Args) != 2 {
		return nil, fmt.Errorf("%s: Key and time must be provided, default seconds, Syntax: TTL.EXPIRE <key> <seconds[s]|minutes[m]|hours[h]|days[d]>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	secondsStr := cmd.C.Args[1]
	// parse argument
	seconds, err := utils.ParseDurationToSeconds(secondsStr)
	if err != nil {
		return nil, fmt.Errorf("%s: Invalid time value provided", errors.InvalidArgsError)
	}

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	success, err := shard.M.Expire(key, seconds)
	if err != nil {
		return nil, err
	}

	var result string
	if success {
		result = "1"
		// Log to WAL if expiration was set successfully
		obj, _ := shard.M.Get(key)
		if obj != nil {
			cmd.SM.Wal().Put(key, obj)
		}
	} else {
		result = "0"
	}

	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(result),
		},
	}, nil
}
