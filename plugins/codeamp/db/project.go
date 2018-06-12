package db_resolver

import (
	"context"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

type ProjectResolver struct {
	model.Project
	model.Environment
	DB *gorm.DB
}

// Features
func (r *ProjectResolver) Features(args *struct{ ShowDeployed *bool }) []*FeatureResolver {
	var rows []model.Feature
	var results []*FeatureResolver

	created := time.Date(0, 0, 0, 0, 0, 0, 0, time.UTC)
	showDeployed := false
	if args.ShowDeployed != nil {
		showDeployed = *args.ShowDeployed
	}

	if !showDeployed {
		var currentRelease model.Release

		if r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").First(&currentRelease).RecordNotFound() {

		} else {
			feature := model.Feature{}
			r.DB.Where("id = ?", currentRelease.HeadFeatureID).First(&feature)
			created = feature.Created
		}
	}

	r.DB.Where("project_id = ? AND ref = ? AND created > ?", r.Project.ID, fmt.Sprintf("refs/heads/%s", r.GitBranch()), created).Order("created desc").Find(&rows)

	for _, feature := range rows {
		results = append(results, &FeatureResolver{DB: r.DB, Feature: feature})
	}

	return results
}

// CurrentRelease
func (r *ProjectResolver) CurrentRelease() (*ReleaseResolver, error) {
	var currentRelease model.Release

	if r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").First(&currentRelease).RecordNotFound() {
		log.InfoWithFields("currentRelease does not exist", log.Fields{
			"state":          transistor.GetState("complete"),
			"project_id":     r.Project.Model.ID,
			"environment_id": r.Environment.Model.ID,
		})
		return &ReleaseResolver{}, fmt.Errorf("CurrentRelease does not exist.")
	}

	return &ReleaseResolver{DB: r.DB, Release: currentRelease}, nil
}

// Releases
func (r *ProjectResolver) Releases(args *struct {
	Params *model.PaginatorInput
}) *ReleaseListResolver {
	var rows []model.Release

	r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").Find(&rows)
	return &ReleaseListResolver{
		DB:             r.DB,
		ReleaseList:    rows,
		PaginatorInput: *args.Params,
	}
}

// Services
func (r *ProjectResolver) Services() []*ServiceResolver {
	var rows []model.Service
	var results []*ServiceResolver

	r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Find(&rows)
	for _, service := range rows {
		results = append(results, &ServiceResolver{DB: r.DB, Service: service})
	}

	return results
}

// Secrets
func (r *ProjectResolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Secret
	var results []*SecretResolver

	r.DB.Select("key, id, created_at, type, project_id, environment_id, deleted_at, is_secret").Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("key, created_at desc").Find(&rows)
	for _, secret := range rows {
		results = append(results, &SecretResolver{DB: r.DB, Secret: secret})
	}

	return results, nil
}

// ProjectExtensions
func (r *ProjectResolver) Extensions() ([]*ProjectExtensionResolver, error) {
	var rows []model.ProjectExtension
	var results []*ProjectExtensionResolver

	r.DB.Where("project_extensions.project_id = ? and project_extensions.environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Joins(`INNER JOIN extensions ON project_extensions.extension_id = extensions.id`).Order(`
		CASE extensions.type
			WHEN 'workflow' THEN 1
			WHEN 'deployment' THEN 2
			ELSE 3
		END, extensions.key ASC`).Find(&rows)

	for _, extension := range rows {
		results = append(results, &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension})
	}

	return results, nil
}

// GitBranch
func (r *ProjectResolver) GitBranch() string {
	var projectSettings model.ProjectSettings

	if r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID.String(), r.Environment.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		return "master"
	} else {
		return projectSettings.GitBranch
	}
}

// ContinuousDeploy
func (r *ProjectResolver) ContinuousDeploy() bool {
	var projectSettings model.ProjectSettings

	if r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID.String(), r.Environment.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		return false
	} else {
		return projectSettings.ContinuousDeploy
	}
}

// Environments
func (r *ProjectResolver) Environments() []*EnvironmentResolver {
	var permissions []model.ProjectEnvironment
	var results []*EnvironmentResolver

	r.DB.Where("project_id = ?", r.Project.ID).Find(&permissions)
	for _, permission := range permissions {
		var environment model.Environment
		r.DB.Where("id = ?", permission.EnvironmentID).Find(&environment)
		results = append(results, &EnvironmentResolver{DB: r.DB, Environment: environment, Project: r.Project})
	}

	return results
}

// Bookmarked
func (r *ProjectResolver) Bookmarked(ctx context.Context) bool {
	var userID string
	var err error

	if userID, err = auth.CheckAuth(ctx, []string{}); err != nil {
		return false
	}

	if r.DB.Where("project_id = ? and user_id = ?", r.Project.Model.ID.String(), userID).First(&model.ProjectBookmark{}).RecordNotFound() {
		return false
	} else {
		return true
	}
}
