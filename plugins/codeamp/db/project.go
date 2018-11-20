package db_resolver

import (
	"context"
	"fmt"

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
func (r *ProjectResolver) Features(args *struct {
	ShowDeployed *bool
	Params       *model.PaginatorInput
}) *FeatureListResolver {

	db := r.DB.Where("project_id = ? AND ref = ?", r.Project.ID, fmt.Sprintf("refs/heads/%s", r.GitBranch()))

	if args.ShowDeployed != nil && *args.ShowDeployed == false {
		var currentRelease model.Release
		if err := r.DB.Where("state = ? and project_id = ? and environment_id = ?", transistor.GetState("complete"), r.Project.Model.ID, r.Environment.Model.ID).Order("created_at desc").First(&currentRelease).Error; err == nil {
			feature := model.Feature{}

			if err := r.DB.Where("id = ?", currentRelease.HeadFeatureID).First(&feature).Error; err == nil {
				db = r.DB.Where("project_id = ? AND ref = ? AND created > ?", r.Project.ID, fmt.Sprintf("refs/heads/%s", r.GitBranch()), feature.Created)
			}
		}
	}

	return &FeatureListResolver{
		PaginatorInput: args.Params,
		DB:             db,
	}
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
		return nil, fmt.Errorf("CurrentRelease does not exist.")
	}

	return &ReleaseResolver{DB: r.DB, Release: currentRelease}, nil
}

// Releases
func (r *ProjectResolver) Releases(args *struct {
	Params *model.PaginatorInput
}) *ReleaseListResolver {
	var db *gorm.DB
	if r.Environment != (model.Environment{}) {
		db = r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID)
	} else {
		db = r.DB.Where("project_id = ?", r.Project.Model.ID)
	}

	return &ReleaseListResolver{
		PaginatorInput: args.Params,
		DB:             db,
	}
}

// Services
func (r *ProjectResolver) Services(args *struct {
	Params *model.PaginatorInput
}) *ServiceListResolver {

	db := r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("name asc")
	return &ServiceListResolver{
		DB:             db,
		PaginatorInput: args.Params,
	}
}

// Secrets
func (r *ProjectResolver) Secrets(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*SecretListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	db := r.DB.Where("project_id = ? and environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Order("key asc")
	return &SecretListResolver{
		DB:             db,
		PaginatorInput: args.Params,
	}, nil
}

// ProjectExtensions
func (r *ProjectResolver) Extensions() ([]*ProjectExtensionResolver, error) {
	var rows []model.ProjectExtension
	var results []*ProjectExtensionResolver

	r.DB.Where("project_extensions.project_id = ? and project_extensions.environment_id = ?", r.Project.Model.ID, r.Environment.Model.ID).Joins(`INNER JOIN extensions ON project_extensions.extension_id = extensions.id`).Order(`
		extensions.type ASC, extensions.key ASC`).Find(&rows)

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
	if err := r.DB.Where("project_id = ?", r.Project.ID).Find(&permissions).Error; err != nil {
		log.Error(err)
	}

	envIDs := make([]string, 0, 0)
	for _, permission := range permissions {
		envIDs = append(envIDs, permission.EnvironmentID.String())
	}

	var environments []model.Environment
	if err := r.DB.Where("id IN (?)", envIDs).Order("name desc").Find(&environments).Error; err != nil {
		log.Error(err.Error())
	}

	var results []*EnvironmentResolver
	for _, environment := range environments {
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
