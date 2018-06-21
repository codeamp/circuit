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

func (r *ServiceResolverQuery) Services(ctx context.Context) ([]*ServiceResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.Service
	var results []*ServiceResolver

	r.DB.Order("created_at desc").Find(&rows)
	for _, service := range rows {
		results = append(results, &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}})
	}

	return results, nil
}
