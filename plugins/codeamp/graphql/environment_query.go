package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Environment Resolver Query
type EnvironmentResolverQuery struct {
	DB *gorm.DB
}

func (r *EnvironmentResolverQuery) Environments(ctx context.Context, args *struct{ ProjectSlug *string }) ([]*EnvironmentResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var environments []model.Environment
	var results []*EnvironmentResolver

	if args.ProjectSlug != nil {
		var project model.Project
		var permissions []model.ProjectEnvironment

		if err := r.DB.Where("slug = ?", *args.ProjectSlug).First(&project).Error; err != nil {
			return nil, err
		}

		r.DB.Where("project_id = ?", project.Model.ID).Find(&permissions)
		for _, permission := range permissions {
			var environment model.Environment
			r.DB.Where("id = ?", permission.EnvironmentID).Find(&environment)
			results = append(results, &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: environment}})
		}

		return results, nil
	}

	r.DB.Order("created_at desc").Find(&environments)
	for _, environment := range environments {
		results = append(results, &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: environment}})
	}

	return results, nil
}
