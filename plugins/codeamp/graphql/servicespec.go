package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// ServiceSpecResolver resolver for ServiceSpec
type ServiceSpecResolver struct {
	DBServiceSpecResolver *db_resolver.ServiceSpecResolver
}

// ID
func (r *ServiceSpecResolver) ID() graphql.ID {
	return graphql.ID(r.DBServiceSpecResolver.ServiceSpec.Model.ID.String())
}

// Name
func (r *ServiceSpecResolver) Name() string {
	return r.DBServiceSpecResolver.ServiceSpec.Name
}

// CpuRequest
func (r *ServiceSpecResolver) CpuRequest() string {
	return r.DBServiceSpecResolver.ServiceSpec.CpuRequest
}

// CpuLimit
func (r *ServiceSpecResolver) CpuLimit() string {
	return r.DBServiceSpecResolver.ServiceSpec.CpuLimit
}

// MemoryRequest
func (r *ServiceSpecResolver) MemoryRequest() string {
	return r.DBServiceSpecResolver.ServiceSpec.MemoryRequest
}

// MemoryLimit
func (r *ServiceSpecResolver) MemoryLimit() string {
	return r.DBServiceSpecResolver.ServiceSpec.MemoryLimit
}

// TerminationGracePeriod
func (r *ServiceSpecResolver) TerminationGracePeriod() string {
	return r.DBServiceSpecResolver.ServiceSpec.TerminationGracePeriod
}

// IsDefault
func (r *ServiceSpecResolver) IsDefault() bool {
	return r.DBServiceSpecResolver.ServiceSpec.IsDefault
}

// Service
func (r *ServiceSpecResolver) Service() *ServiceResolver {
	return &ServiceResolver{DBServiceResolver: r.DBServiceSpecResolver.Service()}
}

// Created
func (r *ServiceSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBServiceSpecResolver.ServiceSpec.Model.CreatedAt}
}

func (r *ServiceSpecResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBServiceSpecResolver.ServiceSpec)
}

func (r *ServiceSpecResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBServiceSpecResolver.ServiceSpec)
}
