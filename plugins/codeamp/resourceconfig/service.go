package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/helpers"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ServiceConfig struct {
	ProjectConfig
	service *model.Service
}

type Service struct {
	*model.Service
}

func CreateProjectServiceConfig(db *gorm.DB, service *model.Service, project *model.Project, env *model.Environment) *ServiceConfig {
	return &ServiceConfig{
		service: service,
		ProjectConfig: ProjectConfig{
			db:          db,
			project:     project,
			environment: env,
		},
	}
}

func (p *ServiceConfig) Import(service *Service) error {
	if p.db == nil || p.project == nil || p.environment == nil {
		return fmt.Errorf(NilDependencyForExportErr, "db, project, environment")
	}

	// check if service with name already exists
	if err := p.db.Where("name = ? and environment_id = ? and project_id = ?", service.Name, p.environment.Model.ID, p.project.Model.ID).Find(&model.Service{}).Error; err == nil {
		return fmt.Errorf(ObjectAlreadyExistsErr, "Service")
	}

	livenessProbeHttpHeaderInputs := []model.HealthProbeHttpHeaderInput{}
	for _, header := range service.LivenessProbe.HttpHeaders {
		livenessProbeHttpHeaderInputs = append(livenessProbeHttpHeaderInputs, model.HealthProbeHttpHeaderInput{
			Name:  header.Name,
			Value: header.Value,
		})
	}

	readinessProbeHttpHeaderInputs := []model.HealthProbeHttpHeaderInput{}
	for _, header := range service.LivenessProbe.HttpHeaders {
		readinessProbeHttpHeaderInputs = append(readinessProbeHttpHeaderInputs, model.HealthProbeHttpHeaderInput{
			Name:  header.Name,
			Value: header.Value,
		})
	}

	// create base service
	serviceInput := model.ServiceInput{
		ProjectID:          p.project.Model.ID.String(),
		Command:            service.Command,
		Name:               service.Name,
		Count:              service.Count,
		Type:               string(service.Type),
		EnvironmentID:      p.environment.Model.ID.String(),
		DeploymentStrategy: &model.DeploymentStrategyInput{},
		LivenessProbe: &model.ServiceHealthProbeInput{
			Type:                &service.LivenessProbe.Type,
			Method:              service.LivenessProbe.Method,
			Scheme:              &service.LivenessProbe.Scheme,
			Path:                &service.LivenessProbe.Path,
			InitialDelaySeconds: &service.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       &service.LivenessProbe.PeriodSeconds,
			TimeoutSeconds:      &service.LivenessProbe.TimeoutSeconds,
			SuccessThreshold:    &service.LivenessProbe.SuccessThreshold,
			FailureThreshold:    &service.LivenessProbe.FailureThreshold,
			HttpHeaders:         &livenessProbeHttpHeaderInputs,
			Port:                &service.LivenessProbe.Port,
		},
		ReadinessProbe: &model.ServiceHealthProbeInput{
			Type:                &service.LivenessProbe.Type,
			Method:              service.LivenessProbe.Method,
			Scheme:              &service.LivenessProbe.Scheme,
			Path:                &service.LivenessProbe.Path,
			InitialDelaySeconds: &service.LivenessProbe.InitialDelaySeconds,
			PeriodSeconds:       &service.LivenessProbe.PeriodSeconds,
			TimeoutSeconds:      &service.LivenessProbe.TimeoutSeconds,
			SuccessThreshold:    &service.LivenessProbe.SuccessThreshold,
			FailureThreshold:    &service.LivenessProbe.FailureThreshold,
			HttpHeaders:         &readinessProbeHttpHeaderInputs,
			Port:                &service.LivenessProbe.Port,
		},
		PreStopHook: &service.PreStopHook,
	}

	_, err := helpers.CreateServiceInDB(p.db, &serviceInput)
	if err != nil {
		return err
	}

	return nil
}

func (s *ServiceConfig) Export() (*Service, error) {
	if s.service == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "service")
	}

	return &Service{
		Service: s.service,
	}, nil
}
