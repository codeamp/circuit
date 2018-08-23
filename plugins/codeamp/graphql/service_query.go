package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Service Resolver Query
type ServiceResolverQuery struct {
	DB *gorm.DB
}

func (r *ServiceResolverQuery) Services(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*ServiceListResolver, error) {

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	return &ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			DB:             r.DB,
			PaginatorInput: args.Params,
		},
	}, nil
}
