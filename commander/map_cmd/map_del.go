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

var mapDeleteCmd = &common.Command{
	Name:   "MAP.DEL",
	Syntax: "MAP.DEL <key> <mapKey>",
	Description: `Deletes a key or a specific field from a map. If only key is provided, deletes the entire map.
					If both key and mapKey are provided, deletes only the specified field from the map.`,
	Example: `
	# Delete a specific field from a map
	localhost:9219> MAP.DEL user-001:devices device-6d6f6sa66d
	Ok

	# Delete an entire map
	localhost:9219> MAP.DEL user-001:settings
	Ok
	`,
	Execute: mapDeleteKey,
}

func init() {
	common.Register(mapDeleteCmd.Name, mapDeleteCmd)
}

func mapDeleteKey(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil || len(cmd.C.Args) > 2 {
		return nil, fmt.Errorf("%s: More then 2 args not supported, Supported args Key or key and mapKey", errors.InvalidArgsError)
	}
	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	if len(cmd.C.Args) == 1 {
		shard.M.Delete(key)
		cmd.SM.Wal().Delete(key)
	}

	if len(cmd.C.Args) == 2 {
		mk := cmd.C.Args[1]
		if !utils.IsValidString(mk) {
			return nil, fmt.Errorf("%s: %s", errors.InvalidMapKeyError, err.Error())
		}
		obj, _ := shard.M.Get(key)
		if obj == nil {
			return &common.CmdResponse{
				D: &comm.Response{
					Result: []byte(""),
				},
			}, nil
		}

		M := make(map[string]any)
		if err := json.Unmarshal(obj.Value, &M); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}

		if store.Map != store.Kind(obj.Kind) {
			return nil, fmt.Errorf("%s: The current value associated with the provided key must be a Map type", errors.InvalidValueError)
		}

		if len(M) == 1 {
			shard.M.Delete(key)
			cmd.SM.Wal().Delete(key)
		} else {
			delete(M, mk)
			obj.Value = utils.ObjectToByte(M)
			cmd.SM.Wal().Put(key, obj)
		}
	}

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
