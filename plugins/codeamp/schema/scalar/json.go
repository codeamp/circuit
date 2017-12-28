package scalar

import (
	"encoding/json"
	"fmt"
)

type Json struct {
	json.RawMessage
}

func (_ Json) ImplementsGraphQLType(name string) bool {
	return name == "Json"
}

func (j *Json) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
	case []interface{}:
		_input, err := json.Marshal(input)
		if err != nil {
			return err
		}
		j.RawMessage = _input
		return nil
	case map[string]interface{}:
		_input, err := json.Marshal(input)
		if err != nil {
			return err
		}
		j.RawMessage = _input
		return nil
	case json.RawMessage:
		j.RawMessage = input
		return nil
	case string:
		j.RawMessage = json.RawMessage([]byte(input))
		return nil
	default:
		return fmt.Errorf("JSON type not matched")
	}
}
