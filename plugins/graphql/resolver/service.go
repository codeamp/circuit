package codeamp_resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Service
type Service struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// ServiceSpecID
	ServiceSpecID uuid.UUID `bson:"serviceSpecID" json:"serviceSpecID" gorm:"type:uuid"`
	// Command
	Command string `json:"command"`
	// Name
	Name string `json:"name"`
	// Type
	Type plugins.Type `json:"type"`
	// Count
	Count string `json:"count"`
	// Ports
	Ports []ServicePort `json:"servicePorts"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
}

type ServicePort struct {
	Model `json:-",inline"`
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"-" gorm:"type:uuid"`
	// Protocol
	Protocol string `json:"protocol"`
	// Port
	Port string `json:"port"`
}

// Service Resolver
type ServiceResolver struct {
	Service
	DB *gorm.DB
}

// ID
func (r *ServiceResolver) ID() graphql.ID {
	return graphql.ID(r.Service.Model.ID.String())
}

// Project
func (r *ServiceResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.Service).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// Command
func (r *ServiceResolver) Command() string {
	return r.Service.Command
}

// Name
func (r *ServiceResolver) Name() string {
	return r.Service.Name
}

// ServiceSpec
func (r *ServiceResolver) ServiceSpec() *ServiceSpecResolver {
	var serviceSpec ServiceSpec

	r.DB.Model(r.Service).Related(&serviceSpec)

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}
}

// Count
func (r *ServiceResolver) Count() string {
	return r.Service.Count
}

// ServicePorts
func (r *ServiceResolver) Ports() ([]*JSON, error) {
	var rows []ServicePort
	var results []*JSON

	r.DB.Where("service_id = ?", r.Service.ID).Order("created_at desc").Find(&rows)

	for _, row := range rows {
		if servicePort, err := json.Marshal(&row); err != nil {
			return results, fmt.Errorf("JSON marshal failed")
		} else {
			results = append(results, &JSON{servicePort})
		}
	}

	return results, nil
}

// Environment
func (r *ServiceResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment Environment

	if r.DB.Where("id = ?", r.Service.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"service": r.Service,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}

// Type
func (r *ServiceResolver) Type() string {
	return string(r.Service.Type)
}

// Created
func (r *ServiceResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Service.Model.CreatedAt}
}

func (r *ServiceResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Service)
}

func (r *ServiceResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Service)
}
