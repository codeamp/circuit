package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Extension Resolver Query
type ReleaseExtensionResolverQuery struct {
	DB *gorm.DB
}

func (r *ReleaseExtensionResolverQuery) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, releaseExtension := range rows {
		results = append(results, &ReleaseExtensionResolver{DBReleaseExtensionResolver: &db_resolver.ReleaseExtensionResolver{DB: r.DB, ReleaseExtension: releaseExtension}})
	}

	return results, nil
}
