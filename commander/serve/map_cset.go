package serve

import (
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var mapCSetCmd = &Command{
	Name:        "MAP.CSET",
	Description: "MAP.CSET clean all existing map keys and value and add new one based on the provided new map key and value",
	Example: `
	localhost:9219> map.set user-001:devices device-6d6f6sa66d '{
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}'
	Ok

	Get Output example 1:
	localhost:9219> map.get user-001:devices
	device-6d6f6sa66d {
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}

	Get Output example 2:
	localhost:9219> map.get user-001:devices device-6d6f6sa66d
	{
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}
	`,
	Execute: mapCSetKey,
}

func init() {
	Register(mapCSetCmd.Name, mapCSetCmd)
}

func mapCSetKey(cmd *Cmd) (*CmdResponse, error) {

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
	var u any

	M[mk] = utils.EnsureUnmarshal(value, &u)

	objBytes := utils.ObjectToByte(M)

	if err = shard.M.Set(key, objBytes); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: objBytes})

	return &CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
