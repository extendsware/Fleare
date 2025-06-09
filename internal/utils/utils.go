package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/parashmaity/fleare/internal/logger"
)

type ClientData struct {
	REQUEST_ID string `json:"requestId"`
	COMMAND    string `json:"command"`
	KEY        string `json:"key"`
	PATH       string `json:"path"`
	BODY       []byte `json:"body"`
}

type Query struct {
	ArrayFilter string
	Field       string
	Value       string
}

func IsValidKey(s string) (bool, error) {

	if len(s) < 1 || len(s) > 200 {
		return false, errors.New("invalid key, must be a string and at least 1 to 200 characters long")
	}
	return true, nil
}

func ObjectToBytes(obj any) ([]byte, error) {
	if obj == nil {
		return nil, nil
	}
	switch v := obj.(type) {
	case []byte:
		return v, nil
	case string:
		return []byte(v), nil
	case fmt.Stringer:
		return []byte(v.String()), nil
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64, bool:
		return fmt.Appendf(nil, "%v", v), nil
	default:
		// Handle structs, maps, slices, etc.
		return json.Marshal(v)
	}
}

func ConvertToNumber(value string) (float64, error) {
	// Try integer first
	if i, err := strconv.ParseInt(value, 10, 64); err == nil {
		return float64(i), nil
	}

	// Try float
	if f, err := strconv.ParseFloat(value, 64); err == nil {
		return f, nil
	}

	// Not a number
	return 0, fmt.Errorf("value %q is not a valid number", value)
}

func ParseQuery(queryString string) (Query, error) {
	var query Query

	// Ensure queryString is not empty
	if queryString == "" {
		return query, errors.New("invalid query format")
	}

	// Split on '$' to extract array filter and field:value pair
	parts := strings.SplitN(queryString, "$", 2)
	query.ArrayFilter = parts[0]

	// If a field:value pair exists, parse it
	if len(parts) == 2 {
		fieldValue := strings.SplitN(parts[1], ":", 2)
		if len(fieldValue) != 2 {
			return query, errors.New("invalid field:value format")
		}
		query.Field, query.Value = fieldValue[0], fieldValue[1]
	}

	return query, nil
}

// parsePath supports $field=val, $field!val, $field~val, =val, !val, ~val
func ParsePath(path string) (field, op, value string) {
	if strings.HasPrefix(path, "$") {
		path = path[1:]
		for _, o := range []string{">=", "<=", ">", "<", ":", "!", "~"} {
			if idx := strings.Index(path, o); idx != -1 {
				return path[:idx], o, path[idx+len(o):]
			}
		}
	} else {
		for _, o := range []string{">=", "<=", ">", "<", ":", "!", "~"} {
			if strings.HasPrefix(path, o) {
				return "", o, strings.TrimPrefix(path, o)
			}
		}
	}
	return "", "", ""
}

// Utility functions can be added here

func IsValidString(s string) bool {

	if len(s) == 0 {
		return false
	}
	return true
}

func CheckType(value interface{}) {
	v := reflect.TypeOf(value)
	fmt.Println(v.String())

}

func MergeMaps(i1, i2 interface{}) (map[string]interface{}, error) {
	map1, ok1 := i1.(map[string]interface{})
	map2, ok2 := i2.(map[string]interface{})

	if !ok1 || !ok2 {
		return nil, errors.New("one or both inputs are not JSON objects")
	}

	// Create a new map to hold the merged result
	merged := make(map[string]interface{})

	// Copy map1 into the merged map
	for key, value := range map1 {
		merged[key] = value
	}

	// Add/overwrite with entries from map2
	for key, value := range map2 {
		merged[key] = value
	}

	return merged, nil
}

func UpdateMap(mainItem, updateItem interface{}) (map[string]interface{}, error) {
	map1, ok1 := mainItem.(map[string]interface{})
	map2, ok2 := updateItem.(map[string]interface{})

	if !ok1 || !ok2 {
		return nil, errors.New("one or both inputs are not JSON objects")
	}

	for key, value := range map2 {
		map1[key] = value
	}
	return map1, nil
}

func FilterList(items []interface{}, key string, value string) []interface{} {
	fmt.Println(key, value)
	var filterItem []interface{}

	for i, item := range items {
		// Type assertion with a check
		v, ok := item.(map[string]interface{})
		if !ok {
			fmt.Printf("Skipping item at index %d: not a map\n", i)
			logger.Warn("Skipping item is not a map", map[string]any{
				"index": i,
			})
			continue
		}
		if fmt.Sprintf("%v", v[key]) == fmt.Sprintf("%v", value) {
			filterItem = append(filterItem, v)
		}
	}
	return filterItem
}

func FilterUpdate(items []interface{}, key string, value string, data interface{}) []interface{} {

	for i, item := range items {
		v := item.(map[string]interface{})
		if fmt.Sprintf("%v", v[key]) == fmt.Sprintf("%v", value) {
			up, err := UpdateMap(v, data)
			if err != nil {
				logger.Error("", err, nil)
			}
			items[i] = up
		}
	}
	return items
}
func ParseJson(jsonString string) (ClientData, error) {
	var clientObject ClientData
	err := json.Unmarshal([]byte(jsonString), &clientObject)
	if err != nil {
		return clientObject, errors.New("Invalid request data format.")
	}
	return clientObject, nil
}

func ObjectToByte(value any) []byte {
	if value == nil {
		return []byte("")
	}
	d, err := json.Marshal(value)
	if err != nil {
		return nil
	}
	return d
}

func EnsureUnmarshal(value string) any {
	var u any
	if err := json.Unmarshal([]byte(value), &u); err != nil {
		return value
	}
	return u
}

func StringToByte(value string) []byte {
	return []byte(value)
}

func ByteToString(data []byte) string {
	return string(data)
}

func ParseInt(v string) (int, bool) {
	num, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return num, true
}
