package graphql_resolver

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// ReleaseExtensionResolver resolver for ReleaseExtension
type ReleaseExtensionResolver struct {
	model.ReleaseExtension
	DB *gorm.DB
}

// ID
func (r *ReleaseExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.ReleaseExtension.Model.ID.String())
}

// Release
func (r *ReleaseExtensionResolver) Release() (*ReleaseResolver, error) {
	release := model.Release{}

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
	extension := model.ProjectExtension{}

	if r.DB.Unscoped().Where("id = ?", r.ReleaseExtension.ProjectExtensionID).Find(&extension).RecordNotFound() {
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
func (r *ReleaseExtensionResolver) Artifacts() model.JSON {
	return model.JSON{r.ReleaseExtension.Artifacts.RawMessage}
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
