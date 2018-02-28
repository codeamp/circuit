package codeamp_resolvers

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// Extension spec
type ExtensionSpec struct {
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

// ExtensionSpecResolver resolver for ExtensionSpec
type ExtensionSpecResolver struct {
	ExtensionSpec
	DB *gorm.DB
}

// ID
func (r *ExtensionSpecResolver) ID() graphql.ID {
	return graphql.ID(r.ExtensionSpec.Model.ID.String())
}

// Name
func (r *ExtensionSpecResolver) Name() string {
	return r.ExtensionSpec.Name
}

// Component
func (r *ExtensionSpecResolver) Component() string {
	return r.ExtensionSpec.Component
}

// Type
func (r *ExtensionSpecResolver) Type() string {
	return string(r.ExtensionSpec.Type)
}

// Key
func (r *ExtensionSpecResolver) Key() string {
	return r.ExtensionSpec.Key
}

// Environment
func (r *ExtensionSpecResolver) Environment() (*EnvironmentResolver, error) {
	environment := Environment{}

	if r.DB.Where("id = ?", r.ExtensionSpec.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.ExtensionSpec.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}

// Config
func (r *ExtensionSpecResolver) Config() JSON {
	return JSON{r.ExtensionSpec.Config.RawMessage}
}

// Created
func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Model.CreatedAt}
}

func (r *ExtensionSpecResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.ExtensionSpec)
}

func (r *ExtensionSpecResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.ExtensionSpec)
}
