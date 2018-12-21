package db_resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
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

	r.DB.Model(&r.Service).Related(&project)

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

// AutoscaleEnabled
func (r *ServiceResolver) AutoscaleEnabled() bool {
	return r.Service.AutoscaleEnabled
}

// ServiceSpec
func (r *ServiceResolver) ServiceSpec() *ServiceSpecResolver {
	var serviceSpec model.ServiceSpec

	if err := r.DB.Where("service_id = ? and type != ?", r.Service.Model.ID, "suggested").First(&serviceSpec).Error; err != nil {
		log.ErrorWithFields(err.Error(), log.Fields{
			"service_id": r.Service.Model.ID,
		})
		return nil
	}

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}
}

// SuggestedServiceSpec
func (r *ServiceResolver) SuggestedServiceSpec() *ServiceSpecResolver {
	var serviceSpec model.ServiceSpec

	if err := r.DB.Where("service_id = ? and type = ?", r.Service.Model.ID, "suggested").Order("created_at desc").First(&serviceSpec).Error; err != nil {
		log.ErrorWithFields(err.Error(), log.Fields{
			"service_id": r.Service.Model.ID,
			"type":       "suggested",
		})
		return nil
	}

	return &ServiceSpecResolver{DB: r.DB, ServiceSpec: serviceSpec}
}

// Count
func (r *ServiceResolver) Count() int32 {
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
func (r *ServiceResolver) DeploymentStrategy() (*model.JSON, error) {
	var deploymentStrategy model.ServiceDeploymentStrategy
	var results model.JSON

	r.DB.Where("service_id = ?", r.Service.ID).First(&deploymentStrategy)

	marshaled, err := json.Marshal(&deploymentStrategy)
	if err != nil {
		return &results, fmt.Errorf("DeploymentStrategy: JSON marshal failed")
	}

	return &model.JSON{marshaled}, nil
}

// LivenessProbe
func (r *ServiceResolver) LivenessProbe() (*model.JSON, error) {
	var livenessProbe model.ServiceHealthProbe

	r.DB.Where("service_id = ? and type = ?", r.Service.ID, string(plugins.GetType("livenessProbe"))).First(&livenessProbe)

	var headers []model.ServiceHealthProbeHttpHeader
	r.DB.Where("health_probe_id = ?", livenessProbe.ID).Find(&headers)

	livenessProbe.HttpHeaders = headers

	marshaled, err := json.Marshal(&livenessProbe)
	if err != nil {
		return &model.JSON{}, fmt.Errorf("LivenessProbe: JSON marshal failed")
	}

	return &model.JSON{marshaled}, nil
}

// ReadinessProbe
func (r *ServiceResolver) ReadinessProbe() (*model.JSON, error) {
	var readinessProbe model.ServiceHealthProbe

	r.DB.Where("service_id = ? and type = ?", r.Service.ID, string(plugins.GetType("readinessProbe"))).First(&readinessProbe)

	var headers []model.ServiceHealthProbeHttpHeader
	r.DB.Where("health_probe_id = ?", readinessProbe.ID).Find(&headers)

	readinessProbe.HttpHeaders = headers

	marshaled, err := json.Marshal(&readinessProbe)
	if err != nil {
		return &model.JSON{}, fmt.Errorf("ReadinessProbe: JSON marshal failed")
	}

	return &model.JSON{marshaled}, nil
}

// PreStopHook
func (r *ServiceResolver) PreStopHook() *string {
	return &r.Service.PreStopHook
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
