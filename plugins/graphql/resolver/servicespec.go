package codeamp_resolvers

import (
	"encoding/json"

	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// ServiceSpec
type ServiceSpec struct {
	Model `json:",inline"`
	// Name
	Name string `json:"name"`
	// CpuRequest
	CpuRequest string `json:"cpuRequest"`
	// CpuLimit
	CpuLimit string `json:"cpuLimit"`
	// MemoryRequest
	MemoryRequest string `json:"memoryRequest"`
	// MemoryLimit
	MemoryLimit string `json:"memoryLimit"`
	// TerminationGracePeriod
	TerminationGracePeriod string `json:"terminationGracePeriod"`
}

// ServiceSpecResolver resolver for ServiceSpec
type ServiceSpecResolver struct {
	ServiceSpec
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
