package map_cmd

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/utils"
)

var mapGetCmd = &common.Command{
	Name:   "MAP.GET",
	Syntax: "MAP.GET <key> [<mapKey>]",
	Description: `Retrieves a map stored at the given key. If only the key is provided, returns the full map.
					If a map key is also provided, returns the value associated with that specific key in the map.
					Useful for accessing structured data such as user devices or settings.`,
	Example: `
	localhost:9219> map.get user-001:devices
	Ok {
	  "device-663abc5352": {
	    "batteryLevel": "75%",
	    "deviceName": "iPhone 14 Pro",
	    "osVersion": "iOS 16"
	  },
	  "device-6d6f6sa66d": {
	    "batteryLevel": "85%",
	    "deviceName": "Pixel 7 Pro",
	    "osVersion": "Android 14"
	  }
	}

	map.get user-001:devices device-663abc5352
	Ok {
	  "batteryLevel": "75%",
	  "deviceName": "iPhone 14 Pro",
	  "osVersion": "iOS 16"
	}
	`,
	Execute: mapGetKey,
}

func init() {
	common.Register(mapGetCmd.Name, mapGetCmd)
}

func mapGetKey(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) > 2 {
		return nil, fmt.Errorf("%s: More then 2 args not supported, Supported args Key or key and mapKey", errors.InvalidArgsError)
	}
	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)

	obj, _ := shard.M.Get(key)

	if obj == nil {
		return &common.CmdResponse{
			D: &comm.Response{
				Result: []byte(""),
			},
		}, nil
	}

	var V []byte = obj.Value

	if len(cmd.C.Args) == 2 {
		mapData := make(map[string]any)
		if err := json.Unmarshal(obj.Value, &mapData); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}

		mk := cmd.C.Args[1]
		if !utils.IsValidString(mk) {
			return nil, fmt.Errorf("%s: %s", errors.InvalidMapKeyError, err.Error())
		}

		if val, ok := mapData[mk]; ok {
			V = utils.ObjectToByte(&val)
		} else {
			V = []byte("")
		}
	}

	return &common.CmdResponse{
		D: &comm.Response{
			Result: V,
		},
	}, nil
}
