package json_cmd

import (
	"encoding/json"
	"fmt"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/config"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var jsonRefSetCmd = &common.Command{
	Name:        "JSON.SETREF",
	Description: ``,

	Syntax: "JSON.SETREF <key> <value> <ref_object>",
	Example: `
	localhost:9219> JSON.SET users:001 '{"name":"John","age":30,"hobbies":["reading","hiking"]}'
	Ok

	localhost:9219> JSON.SETREF orders:OD001 '{"orderId":"orders:OD001","details":"This order is for a new laptop.","status":"pending","trackingNumber":"ABC123","deliveryDate":"2023-06-15","amount":1000.5}' '{"userId":"users:001"}'
	Ok

	localhost:9219> JSON.GET orders:OD001
	Ok {
	  "amount": 1000.5,
	  "deliveryDate": "2023-06-15",
	  "details": "This order is for a new laptop.",
	  "orderId": "order:10002003",
	  "status": "pending",
	  "trackingNumber": "ABC123",
	  "userId": "$ref:users:001"
	}

	localhost:9219> JSON.SETREF orders:OD001 '{"orderId":"orders:OD001","details":"This order is for a new laptop.","status":"pending","trackingNumber":"ABC123","deliveryDate":"2023-06-15","amount":1000.5}' '{"userId":"users:001","productId":"products:001"}'
	Ok

	localhost:9219> JSON.GET orders:OD001
	Ok {
	  "amount": 1000.5,
	  "deliveryDate": "2023-06-15",
	  "details": "This order is for a new laptop.",
	  "orderId": "order:10002003",
	  "productId": "$ref:products:001",
	  "status": "pending",
	  "trackingNumber": "ABC123",
	  "userId": "$ref:users:001"
	}

	localhost:9219> JSON.SET offers:001 '{"offer":{"offerId": "001","code":"MDX50","flat":true}}'
	Ok

	JSON.SETREF orders:OD001 '{"orderId":"orders:OD001","details":"This order is for a new laptop.","status":"pending","trackingNumber":"ABC123","deliveryDate":"2023-06-15","amount":1000.5,"offer":{"name":"Offer one"}}' '{"userId":"users:001","productId":"products:001","offer.offerId":"offers:001"}'
	Ok

	`,
	Execute: jsonRefSetFunc,
}

func init() {
	common.Register(jsonRefSetCmd.Name, jsonRefSetCmd)
}

func jsonRefSetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	if len(cmd.C.Args) != 3 {
		return nil, fmt.Errorf("%s: invalid number of arguments, Syntax: JSON.SETREF <key> <value> <ref_object>", errors.InvalidArgsError)
	}

	key := cmd.C.Args[0]
	value := utils.StringToByte(cmd.C.Args[1])
	refObject := utils.StringToByte(cmd.C.Args[2])

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	if !json.Valid(value) {
		return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "provided value not a valid json object")
	}

	if !json.Valid(refObject) {
		return nil, fmt.Errorf("%s: %s", errors.InvalidValueError, "provided refObject not a valid json object")
	}

	var ref map[string]string
	if err := json.Unmarshal(refObject, &ref); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	var M map[string]interface{}
	if err := json.Unmarshal(value, &M); err != nil {
		return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
	}

	for k, v := range ref {
		if config.GetConfig().Misc.StrictMode {
			s := cmd.SM.GetShardByKey(v)
			obj, _ := s.M.Get(v)
			if obj == nil {
				return nil, fmt.Errorf("%s: %s", errors.InvalidReferenceError, fmt.Sprintf("reference key %s not found", v))
			}
		}
		addNestedKey(M, k, fmt.Sprintf("$ref:%s", v))
		// M[k] = fmt.Sprintf("$ref:%s", v)
	}

	shard := cmd.SM.GetShardByKey(key)

	objBytes := utils.ObjectToByte(M)
	if err = shard.M.Set(key, objBytes, store.JSON); err != nil {
		return nil, err
	}

	cmd.SM.Wal().Put(key, &comm.Object{Value: value, Kind: uint32(store.JSON)})

	return &common.CmdResponse{
		D: &comm.Response{
			Result: []byte(""),
		},
	}, nil
}
