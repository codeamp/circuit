package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Extension Resolver Initializer
type ExtensionResolverInitializer struct {
	DB *gorm.DB
}

func (r *ExtensionResolverInitializer) Extensions(ctx context.Context, args *struct{ EnvironmentID *string }) ([]*ExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Extension
	var results []*ExtensionResolver

	if args.EnvironmentID != nil {
		r.DB.Where("extensions.environment_id = ?", args.EnvironmentID).Order(`
			CASE extensions.type
				WHEN 'workflow' THEN 1
				WHEN 'deployment' THEN 2
				ELSE 3
			END, extensions.key ASC`).Find(&rows)
	} else {
		r.DB.Order(`
			CASE extensions.type
				WHEN 'workflow' THEN 1
				WHEN 'deployment' THEN 2
				ELSE 3
			END, extensions.key ASC`).Find(&rows)
	}

	for _, ext := range rows {
		results = append(results, &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: ext}})
	}

	return results, nil
}
