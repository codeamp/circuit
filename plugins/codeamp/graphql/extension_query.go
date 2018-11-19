package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Extension Resolver Query
type ExtensionResolverQuery struct {
	DB *gorm.DB
}

func (r *ExtensionResolverQuery) Extensions(ctx context.Context, args *struct{ EnvironmentID *string }) ([]*ExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Extension
	var results []*ExtensionResolver

	db := r.DB
	if args.EnvironmentID != nil {
		db = db.Where("extensions.environment_id = ?", args.EnvironmentID)
	}

	db.Order("environment_id asc, name asc").Find(&rows)
	for _, ext := range rows {
		results = append(results, &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: ext}})
	}

	return results, nil
}
