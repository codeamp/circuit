package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Feature Resolver Initializer
type FeatureResolverInitializer struct {
	DB *gorm.DB
}

func (r *FeatureResolverInitializer) Features(ctx context.Context) ([]*FeatureResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Feature
	var results []*FeatureResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, feature := range rows {
		results = append(results, &FeatureResolver{DBFeatureResolver: &db_resolver.FeatureResolver{DB: r.DB, Feature: feature}})
	}

	return results, nil
}
