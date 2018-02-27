package codeamp_resolvers

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/bitstributor/core/plugins/api/utils"
	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type Release struct {
	Model `json:",inline"`
	// State
	State plugins.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// ProjectId
	ProjectId uuid.UUID `json:"projectId" gorm:"type:uuid"`
	// User
	User User
	// UserId
	UserID uuid.UUID `json:"userId" gorm:"type:uuid"`
	// HeadFeatureId
	HeadFeatureID uuid.UUID `json:"headFeatureId" gorm:"type:uuid"`
	// TailFeatureId
	TailFeatureID uuid.UUID `json:"tailFeatureId" gorm:"type:uuid"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"`
	// Snapshot
	Snapshot postgres.Jsonb `json:"snapshot" gorm:"type:jsonb;"`
	// FinishedAt
	FinishedAt time.Time
	// EnvironmentId
	EnvironmentId uuid.UUID `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
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
	artifacts := make(map[string]interface{})

	isAdmin := false
	if _, err := utils.CheckAuth(ctx, []string{"admin"}); err == nil {
		isAdmin = true
	}

	err := json.Unmarshal(r.Release.Artifacts.RawMessage, &artifacts)
	if err != nil {
		log.InfoWithFields(err.Error(), log.Fields{
			"input": r.Release.Artifacts.RawMessage,
		})
		return JSON{}, err
	}

	for key, _ := range artifacts {
		if !isAdmin {
			artifacts[key] = ""
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
	if r.DB.Where("id = ?", r.Release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"releaseId": r.Release.Model.ID,
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
