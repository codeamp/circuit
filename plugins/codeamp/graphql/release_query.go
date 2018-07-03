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

func (r *ReleaseResolverQuery) Releases(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (ReleaseListResolver, error) {
	var query *gorm.DB

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return ReleaseListResolver{}, err
	}

	query = r.DB.Order("created_at desc")

	return ReleaseListResolver{
		DBReleaseListResolver: &db_resolver.ReleaseListResolver{
			Query:          query,
			PaginatorInput: args.Params,
			DB:             r.DB,
		},
	}, nil
}
