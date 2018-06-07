package graphql_resolver

import (
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// EnvironmentResolver resolver for Environment
type EnvironmentResolver struct {
	model.Environment
	model.Project
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

// Projects - get projects permissioned for the environment
func (r *EnvironmentResolver) Projects() []*ProjectResolver {
	var permissions []model.ProjectEnvironment
	var results []*ProjectResolver

	r.DB.Where("environment_id = ?", r.Environment.ID).Find(&permissions)
	for _, permission := range permissions {
		var project model.Project
		if !r.DB.Where("id = ?", permission.ProjectID).First(&project).RecordNotFound() {
			results = append(results, &ProjectResolver{DB: r.DB, Project: project, Environment: r.Environment})
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
