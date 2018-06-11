package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Extension Resolver Query
type ReleaseResolverQuery struct {
	DB *gorm.DB
}

func (r *ReleaseResolverQuery) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Release
	var results []*ReleaseResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, release := range rows {
		results = append(results, &ReleaseResolver{DBReleaseResolver: &db_resolver.ReleaseResolver{DB: r.DB, Release: release}})
	}

	return results, nil
}
