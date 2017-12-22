package scalar

import (
	"encoding/json"
	"fmt"

	"github.com/davecgh/go-spew/spew"
)

type Json struct {
	json.RawMessage
}

func (_ Json) ImplementsGraphQLType(name string) bool {
	return name == "Json"
}

func (j *Json) UnmarshalGraphQL(input interface{}) error {
	switch input := input.(type) {
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
		spew.Dump(input)
		return fmt.Errorf("wrong type")
	}
}
