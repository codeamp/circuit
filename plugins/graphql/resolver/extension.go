package codeamp_resolvers

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

// ProjectExtension spec
type Extension struct {
	Model `json:",inline"`
	// Type
	Type plugins.Type `json:"type"`
	// Key
	Key string `json:"key"`
	// Name
	Name string `json:"name"`
	// Component
	Component string `json:"component"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
	// Config
	Config postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
}

// ExtensionResolver resolver for Extension
type ExtensionResolver struct {
	Extension
	DB *gorm.DB
}

// ID
func (r *ExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.Extension.Model.ID.String())
}

// Name
func (r *ExtensionResolver) Name() string {
	return r.Extension.Name
}

// Component
func (r *ExtensionResolver) Component() string {
	return r.Extension.Component
}

// Type
func (r *ExtensionResolver) Type() string {
	return string(r.Extension.Type)
}

// Key
func (r *ExtensionResolver) Key() string {
	return r.Extension.Key
}

// Environment
func (r *ExtensionResolver) Environment() (*EnvironmentResolver, error) {
	environment := Environment{}

	if r.DB.Where("id = ?", r.Extension.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.Extension.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}

// Config
func (r *ExtensionResolver) Config() JSON {
	return JSON{r.Extension.Config.RawMessage}
}

// Created
func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Extension.Model.CreatedAt}
}

func (r *ExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Extension)
}

func (r *ExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Extension)
}
