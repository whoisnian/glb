package config

import jsonv2 "encoding/json/v2"

// JsonUnmarshal parses the JSON-encoded data and stores the result in the struct.
// The pre-filled struct fields should be correctly kept or overridden.
func JsonUnmarshal(data []byte, pStruct any) error {
	return jsonv2.Unmarshal(data, pStruct)
}
