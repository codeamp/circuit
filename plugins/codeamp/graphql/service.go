package graphql_resolver

import (
	"context"
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
)

// Service Resolver
type ServiceResolver struct {
	DBServiceResolver *db_resolver.ServiceResolver
}

// ID
func (r *ServiceResolver) ID() graphql.ID {
	return graphql.ID(r.DBServiceResolver.Service.Model.ID.String())
}

// Project
func (r *ServiceResolver) Project() *ProjectResolver {
	// return r.DBServiceResolver.Project()
	return nil
}

// Command
func (r *ServiceResolver) Command() string {
	return r.DBServiceResolver.Service.Command
}

// Name
func (r *ServiceResolver) Name() string {
	return r.DBServiceResolver.Service.Name
}

// ServiceSpec
func (r *ServiceResolver) ServiceSpec() *ServiceSpecResolver {
	// return r.DBServiceResolver.ServiceSpec()
	return nil
}

// Count
func (r *ServiceResolver) Count() string {
	return r.DBServiceResolver.Service.Count
}

// ServicePorts
func (r *ServiceResolver) Ports() ([]*model.JSON, error) {
	return r.DBServiceResolver.Ports()
}

// Environment
func (r *ServiceResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	// return r.DBServiceResolver.Environment(ctx)
	return nil, nil
}

// Type
func (r *ServiceResolver) Type() string {
	return string(r.DBServiceResolver.Service.Type)
}

// Created
func (r *ServiceResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBServiceResolver.Service.Model.CreatedAt}
}

func (r *ServiceResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBServiceResolver.Service)
}

func (r *ServiceResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBServiceResolver.Service)
}
