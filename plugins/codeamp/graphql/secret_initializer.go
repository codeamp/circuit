package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Secret Resolver Initializer
type SecretResolverInitializer struct {
	DB *gorm.DB
}

func (r *SecretResolverInitializer) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil, err
	}

	var rows []model.Secret
	var results []*SecretResolver

	r.DB.Where("scope != ?", "project").Order("created_at desc").Find(&rows)
	for _, secret := range rows {
		var secretValue model.SecretValue
		r.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		results = append(results, &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}})
	}

	return results, nil
}
