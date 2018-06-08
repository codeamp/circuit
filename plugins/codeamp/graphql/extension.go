package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
)

// ExtensionResolver resolver for Extension
type ExtensionResolver struct {
	DBExtensionResolver *db_resolver.ExtensionResolver
}

// ID
func (r *ExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.DBExtensionResolver.Extension.Model.ID.String())
}

// Name
func (r *ExtensionResolver) Name() string {
	return r.DBExtensionResolver.Extension.Name
}

// Component
func (r *ExtensionResolver) Component() string {
	return r.DBExtensionResolver.Extension.Component
}

// Type
func (r *ExtensionResolver) Type() string {
	return string(r.DBExtensionResolver.Extension.Type)
}

// Key
func (r *ExtensionResolver) Key() string {
	return r.DBExtensionResolver.Extension.Key
}

// Environment
func (r *ExtensionResolver) Environment() (*EnvironmentResolver, error) {
	// return r.DBExtensionResolver.Environment()
	return nil, nil
}

// Config
func (r *ExtensionResolver) Config() model.JSON {
	return model.JSON{r.DBExtensionResolver.Extension.Config.RawMessage}
}

// Created
func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBExtensionResolver.Extension.Model.CreatedAt}
}

func (r *ExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBExtensionResolver.Extension)
}

func (r *ExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBExtensionResolver.Extension)
}
