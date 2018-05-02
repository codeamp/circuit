package codeamp_resolvers

import (
	"encoding/json"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
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
	Project
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

// ProjectReleases
func (r *EnvironmentResolver) ProjectReleases() []*ReleaseResolver {
	var rows []Release
	var results []*ReleaseResolver

	if r.Project != (Project{}) {
		r.DB.Where("environment_id = ? and project_id = ?", r.Environment.Model.ID, r.Project.Model.ID).Find(&rows)

		for _, row := range rows {
			results = append(results, &ReleaseResolver{DB: r.DB, Release: row})
		}
	}

	return results
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
