package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// ServiceSpec Resolver Query
type ServiceSpecResolverQuery struct {
	DB *gorm.DB
}

func (r *ServiceSpecResolverQuery) ServiceSpecs(ctx context.Context) ([]*ServiceSpecResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []model.ServiceSpec
	var results []*ServiceSpecResolver

	r.DB.Order("name desc").Find(&rows)
	for _, serviceSpec := range rows {
		results = append(results, &ServiceSpecResolver{DBServiceSpecResolver: &db_resolver.ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}})
	}

	return results, nil
}
