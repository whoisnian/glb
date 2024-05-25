package config

import "encoding/json"

// JsonUnmarshal parses the JSON-encoded data and stores the result in the struct.
// The pre-filled struct fields should be correctly kept or overridden.
func JsonUnmarshal(data []byte, pStruct any) error {
	return json.Unmarshal(data, pStruct)
}
