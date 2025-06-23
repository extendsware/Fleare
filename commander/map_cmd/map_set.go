package map_cmd

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var mapSetCmd = &common.Command{
	Name:   "MAP.SET",
	Syntax: "MAP.SET <key> <mapKey> <value>",
	Description: `Stores or updates a key-value pair within a map object stored at the specified key.
					If the map does not exist, it will be created. The <mapKey> is used as the sub-key inside the
					map, and <value> should be a valid JSON object or value. This command enables structured
					data storage under a single top-level key.`,
	Example: `
	localhost:9219> MAP.SET user-001:devices device-6d6f6sa66d '{
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}'
	Ok

	localhost:9219> MAP.SET user-001:devices device-663abc5352 '{
	  "deviceName": "iPhone 14 Pro",
	  "osVersion": "iOS 16",
	  "batteryLevel": "85%"
	}'
	Ok

	localhost:9219> MAP.GET user-001:devices
	Ok {
		device-6d6f6sa66d: {
			deviceName: "Pixel 7 Pro",
			osVersion: "Android 14",
			batteryLevel: "85%"
		},
		device-663abc5352: {
			deviceName: "iPhone 14 Pro",
			osVersion: "iOS 16",
			batteryLevel: "85%"
		}
	}

	localhost:9219> MAP.GET user-001:devices device-663abc5352
	Ok {
		deviceName: "iPhone 14 Pro",
		osVersion: "iOS 16",
		batteryLevel: "85%"
	}
	`,
	Execute: mapSetKey,
}

func init() {
	common.Register(mapSetCmd.Name, mapSetCmd)
}

func mapSetKey(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: Key mapKey, and value must be provided, Syntax: MAP.SET <key> <mapKey> <value>", errors.InvalidArgsError)
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

	obj, err := shard.M.Get(key)
	if err != nil {
		return nil, err
	}

	M := make(map[string]any)
	if obj != nil {
		if err := json.Unmarshal(obj.Value, &M); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}
	}

	if store.Map != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a Map type", errors.InvalidValueError)
	}

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
