package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Feature Resolver Query
type FeatureResolverQuery struct {
	DB *gorm.DB
}

func (r *FeatureResolverQuery) Features(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (FeatureListResolver, error) {
	var query *gorm.DB

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return FeatureListResolver{}, err
	}

	query = r.DB.Order("created_at desc")
	return FeatureListResolver{
		DBFeatureListResolver: &db_resolver.FeatureListResolver{
			DB:             r.DB,
			Query:          query,
			PaginatorInput: args.Params,
		},
	}, nil

}
