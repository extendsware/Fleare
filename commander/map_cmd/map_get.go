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
	Name:        "MAP.GET",
	Description: "MAP.GET key mapKey",
	Example: `
	localhost:9219> map.get
	Ok {
	  "m1": 123.33,
	  "m2": true,
	  "m3": "this is a test message...123",
	  "m4": {
	    "add": {
	      "location": "kolkata"
	    },
	    "name": "parash"
	  }
	}

	localhost:9219> map.get m4
	Ok {
	  "m4": {
	    "add": {
	      "location": "kolkata"
	    },
	    "name": "parash"
	  }
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
