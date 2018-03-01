package codeamp_resolvers

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// ReleaseExtension
type ReleaseExtension struct {
	Model `json:",inline"`
	// ReleaseID
	ReleaseID uuid.UUID `json:"releaseID" gorm:"type:uuid"`
	// FetureHash
	FeatureHash string `json:"featureHash"`
	// ServicesSignature
	ServicesSignature string `json:"servicesSignature"`
	// SecretsSignature
	SecretsSignature string `json:"secretsSignature"`
	// ProjectExtensionID
	ProjectExtensionID uuid.UUID `json:"extensionID" gorm:"type:uuid"`
	// State
	State plugins.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// Type
	Type plugins.Type `json:"type"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"` // captured on workflow success/ fail
	// Finished
	Finished time.Time
}

// ReleaseExtensionResolver resolver for ReleaseExtension
type ReleaseExtensionResolver struct {
	ReleaseExtension
	DB *gorm.DB
}

// ID
func (r *ReleaseExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.ReleaseExtension.Model.ID.String())
}

// Release
func (r *ReleaseExtensionResolver) Release() (*ReleaseResolver, error) {
	release := Release{}

	if r.DB.Where("id = ?", r.ReleaseExtension.ReleaseID.String()).Find(&release).RecordNotFound() {
		log.InfoWithFields("extension not found", log.Fields{
			"id": r.ReleaseExtension.ReleaseID.String(),
		})
		return &ReleaseResolver{DB: r.DB, Release: release}, fmt.Errorf("Couldn't find release")
	}

	return &ReleaseResolver{DB: r.DB, Release: release}, nil
}

// ProjectExtension
func (r *ReleaseExtensionResolver) Extension() (*ProjectExtensionResolver, error) {
	extension := ProjectExtension{}

	if r.DB.Where("id = ?", r.ReleaseExtension.ProjectExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("extension not found", log.Fields{
			"id": r.ReleaseExtension.ProjectExtensionID,
		})
		return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, fmt.Errorf("Couldn't find extension")
	}

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, nil
}

// ServicesSignature
func (r *ReleaseExtensionResolver) ServicesSignature() string {
	return r.ReleaseExtension.ServicesSignature
}

// SecretsSignature
func (r *ReleaseExtensionResolver) SecretsSignature() string {
	return r.ReleaseExtension.SecretsSignature
}

// State
func (r *ReleaseExtensionResolver) State() string {
	return string(r.ReleaseExtension.State)
}

// Type
func (r *ReleaseExtensionResolver) Type() string {
	return string(r.ReleaseExtension.Type)
}

// StateMessage
func (r *ReleaseExtensionResolver) StateMessage() string {
	return r.ReleaseExtension.StateMessage
}

// Artifacts
func (r *ReleaseExtensionResolver) Artifacts() JSON {
	return JSON{r.ReleaseExtension.Artifacts.RawMessage}
}

// Finished
func (r *ReleaseExtensionResolver) Finished() graphql.Time {
	return graphql.Time{Time: r.ReleaseExtension.Finished}
}

// Created
func (r *ReleaseExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ReleaseExtension.Model.CreatedAt}
}

func (r *ReleaseExtensionResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.ReleaseExtension)
}

func (r *ReleaseExtensionResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.ReleaseExtension)
}
