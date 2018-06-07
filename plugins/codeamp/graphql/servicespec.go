package graphql_resolver

import (
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// ServiceSpecResolver resolver for ServiceSpec
type ServiceSpecResolver struct {
	model.ServiceSpec
	DB *gorm.DB
}

// ID
func (r *ServiceSpecResolver) ID() graphql.ID {
	return graphql.ID(r.ServiceSpec.Model.ID.String())
}

// Name
func (r *ServiceSpecResolver) Name() string {
	return r.ServiceSpec.Name
}

// CpuRequest
func (r *ServiceSpecResolver) CpuRequest() string {
	return r.ServiceSpec.CpuRequest
}

// CpuLimit
func (r *ServiceSpecResolver) CpuLimit() string {
	return r.ServiceSpec.CpuLimit
}

// MemoryRequest
func (r *ServiceSpecResolver) MemoryRequest() string {
	return r.ServiceSpec.MemoryRequest
}

// MemoryLimit
func (r *ServiceSpecResolver) MemoryLimit() string {
	return r.ServiceSpec.MemoryLimit
}

// TerminationGracePeriod
func (r *ServiceSpecResolver) TerminationGracePeriod() string {
	return r.ServiceSpec.TerminationGracePeriod
}

// Created
func (r *ServiceSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ServiceSpec.Model.CreatedAt}
}

func (r *ServiceSpecResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.ServiceSpec)
}

func (r *ServiceSpecResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.ServiceSpec)
}
