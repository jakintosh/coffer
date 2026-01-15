package wire

import "encoding/json"

// Response defines the common API response envelope.
type Response struct {
	Data  any    `json:"data,omitempty"`
	Error *Error `json:"error,omitempty"`
}

// Error defines the common API error shape.
type Error struct {
	Message string `json:"message"`
}

// rawResponse is used internally for two-phase JSON decoding.
type rawResponse struct {
	Data  json.RawMessage `json:"data"`
	Error *Error          `json:"error"`
}

// decodeInto decodes a JSON envelope and unmarshals data into dest.
// dest must be a pointer. Returns API error if present.
func decodeInto(body []byte, dest any) (*Error, error) {
	var raw rawResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, err
	}
	if raw.Error != nil && raw.Error.Message != "" {
		return raw.Error, nil
	}
	if dest == nil || len(raw.Data) == 0 || string(raw.Data) == "null" {
		return nil, nil
	}
	return nil, json.Unmarshal(raw.Data, dest)
}
