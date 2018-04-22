package codeamp_resolvers

import (
	"encoding/json"

	"github.com/jinzhu/gorm/dialects/postgres"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// ProjectType
type ProjectType struct {
	Model `json:",inline"`
	// Name (for display)
	Name string `json:"name"`
	// Key (unique string identifier)
	Key string `json:"key"`
	// AvailableExtensions
	AvailableExtensions postgres.Jsonb `json:"availableExtensions"`
}

// AvailableExtension
type AvailableExtension struct {
	// ExtensionID
	ExtensionID string `json:"extensionID"`
	// IsDefault is installed automatically when the project
	// is created
	IsDefault bool `json:"isDefault"`
}

// ProjectTypeResolver resolver for ProjectType
type ProjectTypeResolver struct {
	ProjectType
	DB *gorm.DB
}

// ID
func (r *ProjectTypeResolver) ID() graphql.ID {
	return graphql.ID(r.ProjectType.Model.ID.String())
}

// Name
func (r *ProjectTypeResolver) Name() string {
	return r.ProjectType.Name
}

// Key
func (r *ProjectTypeResolver) Key() string {
	return r.ProjectType.Key
}

// Color
func (r *ProjectTypeResolver) AvailableExtensions() JSON {
	return JSON{r.ProjectType.AvailableExtensions.RawMessage}
}

// Created
func (r *ProjectTypeResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ProjectType.Model.CreatedAt}
}

func (r *ProjectTypeResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.ProjectType)
}

func (r *ProjectTypeResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.ProjectType)
}
