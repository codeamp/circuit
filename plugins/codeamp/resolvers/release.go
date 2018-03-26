package codeamp_resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

type Release struct {
	Model `json:",inline"`
	// State
	State plugins.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// User
	User User
	// UserID
	UserID uuid.UUID `json:"userID" gorm:"type:uuid"`
	// HeadFeatureID
	HeadFeatureID uuid.UUID `json:"headFeatureID" gorm:"type:uuid"`
	// TailFeatureID
	TailFeatureID uuid.UUID `json:"tailFeatureID" gorm:"type:uuid"`
	// Services
	Services postgres.Jsonb `json:"services" gorm:"type:jsonb;"`
	// Secrets
	Secrets postgres.Jsonb `json:"services" gorm:"type:jsonb;"`
	// ProjectExtensions
	ProjectExtensions postgres.Jsonb `json:"extensions" gorm:"type:jsonb;"`
	// EnvironmentID
	EnvironmentID uuid.UUID `json:"environmentID" gorm:"type:uuid"`
	// FinishedAt
	FinishedAt time.Time
}

// ReleaseResolver resolver for Release
type ReleaseResolver struct {
	Release
	DB *gorm.DB
}

// ID
func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.Release.Model.ID.String())
}

// Project
func (r *ReleaseResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.Release).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// User
func (r *ReleaseResolver) User() *UserResolver {
	var user User

	r.DB.Model(r.Release).Related(&user)

	return &UserResolver{DB: r.DB, User: user}
}

// Artifacts
func (r *ReleaseResolver) Artifacts(ctx context.Context) (JSON, error) {
	var artifacts []transistor.Artifact
	var releaseExtensions []ReleaseExtension

	isAdmin := false
	if _, err := CheckAuth(ctx, []string{"admin"}); err == nil {
		isAdmin = true
	}

	r.DB.Where("release_id = ?", r.Model.ID).Find(&releaseExtensions)

	for _, releaseExtension := range releaseExtensions {
		var _artifacts []transistor.Artifact
		err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &_artifacts)
		if err != nil {
			log.InfoWithFields(err.Error(), log.Fields{
				"input": releaseExtension.Artifacts.RawMessage,
			})
		} else {
			for _, artifact := range _artifacts {
				artifacts = append(artifacts, artifact)
			}
		}
	}

	for i, artifact := range artifacts {
		if !isAdmin && artifact.Secret {
			artifacts[i].Value = ""
		}
	}

	marshalledArtifacts, err := json.Marshal(artifacts)
	if err != nil {
		log.InfoWithFields(err.Error(), log.Fields{
			"input": artifacts,
		})
		return JSON{}, err
	}

	return JSON{json.RawMessage(marshalledArtifacts)}, nil
}

// HeadFeature
func (r *ReleaseResolver) HeadFeature() *FeatureResolver {
	var feature Feature
	r.DB.Where("id = ?", r.Release.HeadFeatureID).First(&feature)
	return &FeatureResolver{DB: r.DB, Feature: feature}
}

// TailFeature
func (r *ReleaseResolver) TailFeature() *FeatureResolver {
	var feature Feature

	r.DB.Where("id = ?", r.Release.TailFeatureID).First(&feature)

	return &FeatureResolver{DB: r.DB, Feature: feature}
}

// State
func (r *ReleaseResolver) State() string {
	return string(r.Release.State)
}

// ReleaseExtensions
func (r *ReleaseResolver) ReleaseExtensions() []*ReleaseExtensionResolver {
	var rows []ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.DB.Where("release_id = ?", r.Release.ID).Find(&rows)
	for _, releaseExtension := range rows {
		results = append(results, &ReleaseExtensionResolver{DB: r.DB, ReleaseExtension: releaseExtension})
	}

	return results
}

// StateMessage
func (r *ReleaseResolver) StateMessage() string {
	return r.Release.StateMessage
}

// Environment
func (r *ReleaseResolver) Environment() (*EnvironmentResolver, error) {
	var environment Environment
	if r.DB.Where("id = ?", r.Release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"releaseID": r.Release.Model.ID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}

// Created
func (r *ReleaseResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Release.Model.CreatedAt}
}

func (r *ReleaseResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Release)
}

func (r *ReleaseResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Release)
}
