package serve

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var mapSetCmd = &Command{
	Name:        "MAP.SET",
	Description: "MAP.SET key mapOnKey value",
	Example: `
	localhost:9219> map.set user-001:devices device-6d6f6sa66d '{
	  "deviceName": "Pixel 7 Pro",
	  "osVersion": "Android 14",
	  "batteryLevel": "85%"
	}'
	Ok
	`,
	Execute: mapSetKey,
}

func init() {
	Register(mapSetCmd.Name, mapSetCmd)
}

func mapSetKey(cmd *Cmd) (*CmdResponse, error) {

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

	obj, err := shard.M.Get(key)
	if err != nil {
		return nil, err
	}

	M := make(map[string]any)
	var u any
	if obj != nil {
		if err := json.Unmarshal(obj.Value, &M); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}
	}
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
