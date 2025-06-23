package map_cmd

import (
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var mapCSetCmd = &common.Command{
	Name:   "MAP.CSET",
	Syntax: "MAP.CSET <key> <mapKey> <value>",
	Description: `Sets or updates a key-value pair inside a map stored at the specified key.
					The entire map is overwritten with a new map containing only the provided mapKey and value.
					This effectively replaces any existing data stored under the given key with a single-entry map.`,
	Example: `
	localhost:9219> MAP.CSET user-001:devices device-6d6f6sa66d '{
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}'
	Ok

	localhost:9219> MAP.GET user-001:devices
	Ok {
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}

	localhost:9219> MAP.CSET user-001:devices device-663abc5352 '{
	  "deviceName": "iPhone 14 Pro",
	  "osVersion": "iOS 16",
	  "batteryLevel": "75%"
	}'

	localhost:9219> MAP.GET user-001:devices
	Ok {
		"deviceName": "iPhone 14 Pro",
		"osVersion": "iOS 16",
		"batteryLevel": "75%"
	}
	`,
	Execute: mapCSetKey,
}

func init() {
	common.Register(mapCSetCmd.Name, mapCSetCmd)
}

func mapCSetKey(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: Key mapKey, and value must be provided", errors.InvalidArgsError)
	}
	key := cmd.C.Args[0]
	mk := cmd.C.Args[1]
	value := cmd.C.Args[2]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	if !utils.IsValidString(mk) {
		return nil, fmt.Errorf("%s: %s", errors.InvalidMapKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	shard.M.Delete(key)

	M := make(map[string]any)

	M[mk] = utils.EnsureUnmarshal(value)

	objBytes := utils.ObjectToByte(M)

	if err = shard.M.Set(key, objBytes, store.Map); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes, Kind: uint32(store.Map)})

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
