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
}) (*SecretListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	db := r.DB.Where("scope != ?", "project").Order("environment_id desc, key asc, scope asc")
	return &SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			DB:             db,
			PaginatorInput: args.Params,
		},
	}, nil
}

func (r *SecretResolverQuery) Secret(ctx context.Context, args *struct {
	ID *string
}) (*SecretResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	resolver := SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB}}

	r.DB.Where("id = ?", args.ID).First(&resolver.DBSecretResolver.Secret)
	return &resolver, nil
}

// ExportSecrets returns a list of all secrets for a given project and environment in a YAML string format
func (r *SecretResolverQuery) ExportSecrets(ctx context.Context, args *struct{ Params *model.ExportSecretsInput }) (string, error) {
	out := ""

	return out, nil
}
