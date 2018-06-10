package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// ProjectExtension Resolver Initializer
type ProjectExtensionResolverInitializer struct {
	DB *gorm.DB
}

func (r *ProjectExtensionResolverInitializer) ProjectExtensions(ctx context.Context) ([]*ProjectExtensionResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ProjectExtension
	var results []*ProjectExtensionResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, extension := range rows {
		results = append(results, &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}})
	}

	return results, nil
}
