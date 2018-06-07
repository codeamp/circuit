package model

import (
	"encoding/json"
	"fmt"
)

// JSON
type JSON struct {
	json.RawMessage
}

func (r *JSON) ImplementsGraphQLType(name string) bool {
	return name == "JSON"
}

func (r *JSON) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case []interface{}:
		_input, err := json.Marshal(input)
		if err != nil {
			return err
		}
		r.RawMessage = _input
		return nil
	case map[string]interface{}:
		_input, err := json.Marshal(input)
		if err != nil {
			return err
		}
		r.RawMessage = _input
		return nil
	case json.RawMessage:
		r.RawMessage = input
		return nil
	case string:
		r.RawMessage = json.RawMessage([]byte(input))
		return nil
	default:
		return fmt.Errorf("JSON type not matched")
	}
}
