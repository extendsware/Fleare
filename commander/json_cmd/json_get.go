package json_cmd

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/parashmaity/fleare/commander/common"
	"github.com/parashmaity/fleare/internal/comm"
	errors "github.com/parashmaity/fleare/internal/errors"
	"github.com/parashmaity/fleare/internal/shard"
	"github.com/parashmaity/fleare/internal/store"
	"github.com/parashmaity/fleare/internal/utils"
)

var jsonGetCmd = &common.Command{
	Name: "JSON.GET",
	Description: `Fetches a JSON value associated with a given key. Optionally, you can specify a JSON path to retrieve a nested value inside the JSON object.
						If no path is specified, the entire JSON object is returned.

						- If the key does not exist, an empty response is returned.
						- If the value associated with the key is not a valid JSON object, an error is returned.
						- If the provided path does not exist or is invalid, an error is returned.`,

	Syntax: "JSON.GET <key> [<path>] [<ref_field>]",
	Example: `
	localhost:9219> JSON.SET myObj '{"name":"John Doe","age":30,"isActive":true,"hobbies":["reading","hiking"],"address":{"street":"123 Main St","city":"New York"}}'
	Ok

	localhost:9219> JSON.GET myObj
	Ok {
		"address": {
			"city": "New York",
			"street": "123 Main St"
		},
		"age": 30,
		"hobbies": [
			"reading",
			"hiking"
		],
		"isActive": true,
		"name": "John Doe"
	}

	localhost:9219> JSON.GET myObj address
	Ok {
		"city": "New York",
		"street": "123 Main St"
	}

	localhost:9219> JSON.GET myObj address.city
	Ok "New York"

	localhost:9219> JSON.GET myObj hobbies
	Ok ["reading","hiking"]
	`,
	Execute: jsonGetFunc,
}

func init() {
	common.Register(jsonGetCmd.Name, jsonGetCmd)
}

func jsonGetFunc(cmd *common.Cmd) (*common.CmdResponse, error) {

	if cmd.C.Args == nil {
		return nil, fmt.Errorf("%s: Key must be provided", errors.InvalidKeyError)
	}

	key := cmd.C.Args[0]

	valid, err := utils.IsValidKey(key)
	if !valid {
		return nil, fmt.Errorf("%s: %s", errors.InvalidKeyError, err.Error())
	}

	shard := cmd.SM.GetShardByKey(key)
	obj, _ := shard.M.Get(key)

	if obj == nil {
		return sentResponse([]byte(""), nil)
	}

	if store.JSON != store.Kind(obj.Kind) {
		return nil, fmt.Errorf("%s: The current value associated with the provided key must be a json object", errors.InvalidValueError)
	}

	if len(cmd.C.Args) == 1 {
		return sentResponse(obj.Value, nil)
	} else {
		var M map[string]interface{}
		path := cmd.C.Args[1]
		if err := json.Unmarshal(obj.Value, &M); err != nil {
			return nil, fmt.Errorf("%s: %s", errors.InvalidCharacterError, err.Error())
		}

		d, _ := readNestedKey(M, path)
		if d == nil {
			return sentResponse([]byte(""), nil)
		}

		if len(cmd.C.Args) > 2 {
			refRead(d.(map[string]interface{}), cmd.C.Args[2:], cmd.SM)
		}

		v := utils.ObjectToByte(d)
		return sentResponse(v, nil)
	}
}

// ref can be single and nested ["userId", "productId", "offer.offerId"]
// check if the ref keys exist in the json object also nested
//
//	input obj = {
//	  "amount": 1000.5,
//	  "deliveryDate": "2023-06-15",
//	  "details": "This order is for a new laptop.",
//	  "offer": {
//	    "name": "Offer one",
//	    "offerId": "$ref:offers:001"
//	  },
//	  "orderId": "orders:OD001",
//	  "productId": "$ref:products:001",
//	  "status": "pending",
//	  "trackingNumber": "ABC123",
//	  "userId": "$ref:users:001"
//	}
//
//	output obj = {
//	  "amount": 1000.5,
//	  "deliveryDate": "2023-06-15",
//	  "details": "This order is for a new laptop.",
//	  "offer": {
//	    "name": "Offer one",
//	    "offerId": "$ref:offers:001"
//		"_offerId": {
//	  	"id": "offers:001",
//	  	"name": "Offer one",
//	  	"code": "MDX50",
//	  	"flat": true
//		}
//	},
//		  "orderId": "orders:OD001",
//		  "productId": "$ref:products:001",
//	   "_productId": {
//	     "id": "products:001",
//	     "name": "Laptop",
//	     "price": 1000.5
//	   },
//		  "status": "pending",
//		  "trackingNumber": "ABC123",
//		  "userId": "$ref:users:001"
//		  "_userId": {
//	      "id": "users:001",
//	  	 "name": "John Doe",
//	  	 "email": "john.doe@example.com"
//		  }
//
//		}
func refRead(obj map[string]interface{}, ref []string, sm *shard.ShardManager) error {
	for _, refPath := range ref {
		if err := resolveReference(obj, refPath, sm); err != nil {
			return err
		}
	}
	return nil
}

func resolveReference(obj map[string]interface{}, refPath string, sm *shard.ShardManager) error {
	// Split the path to handle nested references (e.g., "offer.offerId")
	keys := strings.Split(refPath, ".")

	// Navigate to the correct location in the object
	current := obj
	for i, key := range keys {
		if i == len(keys)-1 {
			// This is the final key that should contain the reference
			val, exists := current[key]
			if !exists {
				return fmt.Errorf("reference key not found: %s", refPath)
			}

			// Check if it's a reference value (starts with "$ref:")
			refValue, ok := val.(string)
			if !ok || !strings.HasPrefix(refValue, "$ref:") {
				// Not a reference, skip
				return nil
			}

			// Extract the actual key from the reference
			actualKey := strings.TrimPrefix(refValue, "$ref:")

			// Fetch the referenced object from the store
			shard := sm.GetShardByKey(actualKey)
			refObj, _ := shard.M.Get(actualKey)
			if refObj == nil {
				return fmt.Errorf("referenced key not found: %s", actualKey)
			}

			// Parse the referenced JSON object
			var refData map[string]interface{}
			if err := json.Unmarshal(refObj.Value, &refData); err != nil {
				return fmt.Errorf("failed to parse referenced object: %s", err.Error())
			}

			// Add the resolved reference with "_" prefix
			refFieldName := "_" + key
			current[refFieldName] = refData

		} else {
			// Navigate deeper into the nested structure
			val, exists := current[key]
			if !exists {
				return fmt.Errorf("path not found: %s", strings.Join(keys[:i+1], "."))
			}

			nextMap, ok := val.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid path structure at: %s", strings.Join(keys[:i+1], "."))
			}
			current = nextMap
		}
	}

	return nil
}

func readNestedKey(obj map[string]interface{}, path string) (interface{}, error) {
	if path == "" {
		return obj, nil
	}

	keys := strings.Split(path, ".")

	current := obj
	for i, key := range keys {
		val, exists := current[key]
		if !exists {
			return nil, fmt.Errorf("key not found: %s", key)
		}

		if i == len(keys)-1 {
			// Return pointer to the final value so it can be modified
			return val, nil
		}

		// Expecting a nested map for further traversal
		nextMap, ok := val.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid path: %s", path)
		}
		current = nextMap
	}
	return nil, fmt.Errorf("unexpected end of path")
}

func sentResponse(data []byte, err error) (*common.CmdResponse, error) {
	if err != nil {
		return nil, err
	}
	return &common.CmdResponse{
		D: &comm.Response{
			Result: data,
		},
	}, err
}
