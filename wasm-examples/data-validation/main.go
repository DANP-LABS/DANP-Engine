package main

import (
	"encoding/json"
	"fmt"

	"github.com/extism/go-pdk"
)

//export validate_data
func validate_data() int32 {
	// Read the input JSON string from the host
	input_string := pdk.InputString()

	var data map[string]interface{}
	err := json.Unmarshal([]byte(input_string), &data)
	if err != nil {
		// Return error message if JSON is invalid
		error_msg := fmt.Sprintf(`{"valid": false, "error": "invalid JSON: %s"}`, err.Error())
		pdk.OutputString(error_msg)
		return 1 // Indicate failure
	}

	// Check if the "signature" key exists
	if _, ok := data["signature"]; !ok {
		error_msg := `{"valid": false, "error": "missing 'signature' key"}`
		pdk.OutputString(error_msg)
		return 1 // Indicate failure
	}

	// If validation is successful
	success_msg := `{"valid": true}`
	pdk.OutputString(success_msg)
	return 0 // Indicate success
}

func main() {}