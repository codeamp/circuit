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

// ProjectExtension
type ProjectExtension struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// ExtensionID
	ExtensionID uuid.UUID `json:"extID" gorm:"type:uuid"`
	// State
	State plugins.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"`
	// Config
	Config postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
	// CustomConfig
	CustomConfig postgres.Jsonb `json:"customConfig" gorm:"type:jsonb;not null"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
}

// ProjectExtensionResolver resolver for ProjectExtension
type ProjectExtensionResolver struct {
	ProjectExtension
	DB *gorm.DB
}

// ID
func (r *ProjectExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.ProjectExtension.Model.ID.String())
}

// Project
func (r *ProjectExtensionResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.ProjectExtension).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// Extension
func (r *ProjectExtensionResolver) Extension() *ExtensionResolver {
	var ext Extension

	r.DB.Model(r.ProjectExtension).Related(&ext)

	return &ExtensionResolver{DB: r.DB, Extension: ext}
}

// Artifacts
func (r *ProjectExtensionResolver) Artifacts() JSON {
	return JSON{r.ProjectExtension.Artifacts.RawMessage}
}

// Config
func (r *ProjectExtensionResolver) Config() JSON {
	return JSON{r.ProjectExtension.Config.RawMessage}
}

// CustomConfig
func (r *ProjectExtensionResolver) CustomConfig() JSON {
	return JSON{r.ProjectExtension.CustomConfig.RawMessage}
}

// State
func (r *ProjectExtensionResolver) State() string {
	return string(r.ProjectExtension.State)
}

// StateMessage
func (r *ProjectExtensionResolver) StateMessage() string {
	return r.ProjectExtension.StateMessage
}

// Environment
func (r *ProjectExtensionResolver) Environment() (*EnvironmentResolver, error) {
	var environment Environment
	if r.DB.Where("id = ?", r.ProjectExtension.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}

// Created
func (r *ProjectExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ProjectExtension.Model.CreatedAt}
}

func (r *ProjectExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.ProjectExtension)
}

func (r *ProjectExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.ProjectExtension)
}
