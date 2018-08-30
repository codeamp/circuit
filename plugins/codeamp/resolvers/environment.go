package codeamp_resolvers

import (
	"encoding/json"

	"github.com/jinzhu/gorm"
	graphql "github.com/graph-gophers/graphql-go"
)

// Environment Environment
type Environment struct {
	Model `json:",inline"`
	// Name
	Name string `json:"name"`
	// Key
	Key string `json:"key"`
	// Is Default
	IsDefault bool `json:"isDefault"`
	// Color
	Color string `json:"color"`
}

// EnvironmentResolver resolver for Environment
type EnvironmentResolver struct {
	Environment
	DB *gorm.DB
}

// ID
func (r *EnvironmentResolver) ID() graphql.ID {
	return graphql.ID(r.Environment.Model.ID.String())
}

// Name
func (r *EnvironmentResolver) Name() string {
	return r.Environment.Name
}

// Color
func (r *EnvironmentResolver) Color() string {
	return r.Environment.Color
}

// Key
func (r *EnvironmentResolver) Key() string {
	return r.Environment.Key
}

// Is Default
func (r *EnvironmentResolver) IsDefault() bool {
	return r.Environment.IsDefault
}

// Created
func (r *EnvironmentResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Environment.Model.CreatedAt}
}

func (r *EnvironmentResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Environment)
}

func (r *EnvironmentResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Environment)
}
