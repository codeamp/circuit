package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Environment Resolver Query
type EnvironmentResolverQuery struct {
	DB *gorm.DB
}

func (r *EnvironmentResolverQuery) Environments(ctx context.Context, args *struct{ ProjectSlug *string }) ([]*EnvironmentResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	log.Warn("ENVIRONMENTS QUERY")

	var environments []model.Environment
	var results []*EnvironmentResolver

	if args.ProjectSlug != nil {
		var envIDs []uuid.UUID
		var project model.Project
		var permissions []model.ProjectEnvironment

		if err := r.DB.Where("slug = ?", *args.ProjectSlug).Find(&project).Error; err != nil {
			return nil, err
		}

		if err := r.DB.Where("project_id = ?", project.Model.ID).Find(&permissions).Error; err != nil {
			log.Error(err)
		}

		for _, permission := range permissions {
			envIDs = append(envIDs, permission.EnvironmentID)
		}

		if err := r.DB.Where("id IN (?)", envIDs).Order("name asc").Find(&environments).Error; err != nil {
			log.Error(err)
		}
	} else {
		if err := r.DB.Order("name asc").Find(&environments).Error; err != nil {
			return nil, err
		}
	}

	for _, environment := range environments {
		results = append(results, &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: environment}})
	}

	return results, nil
}
