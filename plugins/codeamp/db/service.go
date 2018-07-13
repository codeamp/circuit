package db_resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
)

type ServiceResolver struct {
	model.Service
	DB *gorm.DB
}

// Project
func (r *ServiceResolver) Project() *ProjectResolver {
	var project model.Project

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
	var serviceSpec model.ServiceSpec

	r.DB.Model(r.Service).Related(&serviceSpec)

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}
}

// Count
func (r *ServiceResolver) Count() string {
	return r.Service.Count
}

// ServicePorts
func (r *ServiceResolver) Ports() ([]*model.JSON, error) {
	var rows []model.ServicePort
	var results []*model.JSON

	r.DB.Where("service_id = ?", r.Service.ID).Order("created_at desc").Find(&rows)

	for _, row := range rows {
		if servicePort, err := json.Marshal(&row); err != nil {
			return results, fmt.Errorf("JSON marshal failed")
		} else {
			results = append(results, &model.JSON{servicePort})
		}
	}

	return results, nil
}

// DeploymentStrategy
func (r *ServiceResolver) DeploymentStrategy() (model.JSON, error) {
	var deploymentStrategy model.ServiceDeploymentStrategy
	var results model.JSON

	r.DB.Where("service_id = ?", r.Service.ID).First(&deploymentStrategy)

	marshaled, err := json.Marshal(&deploymentStrategy)
	if err != nil {
		return results, fmt.Errorf("DeploymentStrategy: JSON marshal failed")
	}

	return model.JSON{marshaled}, nil
}

// Environment
func (r *ServiceResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var environment model.Environment

	if r.DB.Where("id = ?", r.Service.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"service": r.Service,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}
