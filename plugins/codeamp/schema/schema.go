package schema

import "github.com/codeamp/circuit/assets"

func Schema() (string, error) {
	schema, err := assets.Asset("plugins/codeamp/schema/schema.graphql")
	if err != nil {
		return "", err
	}

	return string(schema), nil
}
