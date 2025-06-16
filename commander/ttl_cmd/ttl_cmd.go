package ttl_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	"github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var ttlCmd = &common.Command{
	Name: "TTL",
	Description: `Returns the remaining time to live (in seconds) of a key. If the key does not exist, it returns -2.
					If the key exists but has no associated expiration, it returns -1.`,
	Syntax: "TTL <key>",
	Example: `
	localhost:9219> TTL.EXPIRE mykey 120
	OK 1

	localhost:9219> TTL mykey
	Ok 85

	localhost:9219> TTL nonexistent
	Ok -2

	localhost:9219> TTL keExists
	Ok -1
	`,
	Execute: ttlKey,
}

func init() {
	common.Register(ttlCmd.Name, ttlCmd)
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

	result := fmt.Sprintf("%d", ttl)
	return &common.CmdResponse{
		ClientID: cmd.ClientID,
		D: &comm.Response{
			ClientId: cmd.ClientID,
			Result:   []byte(result),
		},
	}, nil
}
