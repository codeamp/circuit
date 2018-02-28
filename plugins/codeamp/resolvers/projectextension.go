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

// Extension
type Extension struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// ExtensionSpecID
	ExtensionSpecID uuid.UUID `json:"extensionSpecID" gorm:"type:uuid"`
	// State
	State plugins.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"`
	// Config
	Config postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
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

// Project
func (r *ExtensionResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.Extension).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// ExtensionSpec
func (r *ExtensionResolver) ExtensionSpec() *ExtensionSpecResolver {
	var extensionSpec ExtensionSpec

	r.DB.Model(r.Extension).Related(&extensionSpec)

	return &ExtensionSpecResolver{DB: r.DB, ExtensionSpec: extensionSpec}
}

// Artifacts
func (r *ExtensionResolver) Artifacts() JSON {
	return JSON{r.Extension.Artifacts.RawMessage}
}

// Config
func (r *ExtensionResolver) Config() JSON {
	return JSON{r.Extension.Config.RawMessage}
}

// State
func (r *ExtensionResolver) State() string {
	return string(r.Extension.State)
}

// StateMessage
func (r *ExtensionResolver) StateMessage() string {
	return r.Extension.StateMessage
}

// Environment
func (r *ExtensionResolver) Environment() (*EnvironmentResolver, error) {
	var environment Environment
	if r.DB.Where("id = ?", r.Extension.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.Extension.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
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
