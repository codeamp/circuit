package codeamp_resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// Project
type Project struct {
	Model `json:",inline"`
	// Name
	Name string `json:"name"`
	// Slug
	Slug string `json:"slug"`
	// Repository
	Repository string `json:"repository"`
	// Secret
	Secret string `json:"-"`
	// GitUrl
	GitUrl string `json:"GitUrl"`
	// GitProtocol
	GitProtocol string `json:"GitProtocol"`
	// RsaPrivateKey
	RsaPrivateKey string `json:"-"`
	// RsaPublicKey
	RsaPublicKey string `json:"rsaPublicKey"`
}

// Project settings
type ProjectSettings struct {
	Model `json:"inline"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// GitBranch
	GitBranch string `json:"gitBranch"`
}

// ProjectResolver resolver for Project
type ProjectResolver struct {
	Project
	Environment
	DB *gorm.DB
}

// ID
func (r *ProjectResolver) ID() graphql.ID {
	return graphql.ID(r.Project.Model.ID.String())
}

// Name
func (r *ProjectResolver) Name() string {
	return r.Project.Name
}

// Slug
func (r *ProjectResolver) Slug() string {
	return r.Project.Slug
}

// Repository
func (r *ProjectResolver) Repository() string {
	return r.Project.Repository
}

// Secret
func (r *ProjectResolver) Secret() string {
	return r.Project.Secret
}

// GitUrl
func (r *ProjectResolver) GitUrl() string {
	return r.Project.GitUrl
}

// GitProtocol
func (r *ProjectResolver) GitProtocol() string {
	return r.Project.GitProtocol
}

// RsaPrivateKey
func (r *ProjectResolver) RsaPrivateKey() string {
	return r.Project.RsaPrivateKey
}

// RsaPublicKey
func (r *ProjectResolver) RsaPublicKey() string {
	return r.Project.RsaPublicKey
}

// Features
func (r *ProjectResolver) Features() []*FeatureResolver {
	var rows []Feature
	var results []*FeatureResolver

	r.DB.Where("project_id = ? and ref = ?", r.Project.ID, fmt.Sprintf("refs/heads/%s", r.GitBranch())).Order("created desc").Find(&rows)

	for _, feature := range rows {
		results = append(results, &FeatureResolver{DB: r.DB, Feature: feature})
	}

	return results
}

// CurrentRelease
func (r *ProjectResolver) CurrentRelease() (*ReleaseResolver, error) {
	var currentRelease Release

	if r.DB.Where("state = ? and project_id = ? and environment_id = ?", plugins.GetState("complete"), r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").First(&currentRelease).RecordNotFound() {
		log.InfoWithFields("currentRelease does not exist", log.Fields{
			"state":          plugins.GetState("complete"),
			"project_id":     r.Project.Model.ID,
			"environment_id": r.Environment.Model.ID,
		})
		return &ReleaseResolver{}, fmt.Errorf("CurrentRelease does not exist.")
	}

	return &ReleaseResolver{DB: r.DB, Release: currentRelease}, nil
}

// Releases
func (r *ProjectResolver) Releases() []*ReleaseResolver {
	var rows []Release
	var results []*ReleaseResolver

	r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").Find(&rows)
	for _, release := range rows {
		results = append(results, &ReleaseResolver{DB: r.DB, Release: release})
	}

	return results
}

// Services
func (r *ProjectResolver) Services() []*ServiceResolver {
	var rows []Service
	var results []*ServiceResolver

	r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Find(&rows)
	for _, service := range rows {
		results = append(results, &ServiceResolver{DB: r.DB, Service: service})
	}

	return results
}

// Secrets
func (r *ProjectResolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	if _, err := CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []Secret
	var results []*SecretResolver

	r.DB.Select("key, id, created_at, type, project_id, environment_id, deleted_at, is_secret").Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("key, created_at desc").Find(&rows)
	for _, secret := range rows {
		results = append(results, &SecretResolver{DB: r.DB, Secret: secret})
	}

	return results, nil
}

// ProjectExtensions
func (r *ProjectResolver) Extensions() ([]*ProjectExtensionResolver, error) {
	var rows []ProjectExtension
	var results []*ProjectExtensionResolver

	r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Find(&rows)
	for _, extension := range rows {
		results = append(results, &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension})
	}

	return results, nil
}

// GitBranch
func (r *ProjectResolver) GitBranch() string {
	var projectSettings ProjectSettings

	if r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID.String(), r.Environment.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		return "master"
	} else {
		return projectSettings.GitBranch
	}
}

// Created
func (r *ProjectResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Project.Model.CreatedAt}
}

func (r *ProjectResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Project)
}

func (r *ProjectResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Project)
}
