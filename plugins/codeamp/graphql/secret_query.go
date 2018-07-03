package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Secret Resolver Query
type SecretResolverQuery struct {
	DB *gorm.DB
}

func (r *SecretResolverQuery) Secrets(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (SecretListResolver, error) {
	var query *gorm.DB

	if _, err := auth.CheckAuth(ctx, []string{"admin"}); err != nil {
		return SecretListResolver{}, err
	}

	query = r.DB.Where("scope != ?", "project").Order("created_at desc")

	return SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			DB:             r.DB,
			Query:          query,
			PaginatorInput: args.Params,
		},
	}, nil
}
