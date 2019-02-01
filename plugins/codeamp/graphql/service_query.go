package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

// Service Resolver Query
type ServiceResolverQuery struct {
	DB *gorm.DB
}

func (r *ServiceResolverQuery) Services(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*ServiceListResolver, error) {

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	db := r.DB.Order("name asc")
	return &ServiceListResolver{
		DBServiceListResolver: &db_resolver.ServiceListResolver{
			DB:             db,
			PaginatorInput: args.Params,
		},
	}, nil
}

// ExportServices returns the set of all services in a YAML formatted string
func (r *ServiceResolverQuery) ExportServices(args *struct{ Params *model.ExportServicesInput }) (string, error) {
	project := model.Project{}
	env := model.Environment{}
	services := []model.Service{}

	if err := r.DB.Where("id = ?", args.Params.ProjectID).First(&project).Error; err != nil {
		return "", err
	}

	if err := r.DB.Where("id = ?", args.Params.EnvironmentID).First(&env).Error; err != nil {
		return "", err
	}

	if err := r.DB.Where("project_id = ? and environment_id = ?", project.Model.ID, env.Model.ID).Find(&services).Error; err != nil {
		return "", err
	}

	outServices := []model.Service{}
	for _, service := range services {
		// Get ports
		ports := []model.ServicePort{}
		if err := r.DB.Where("service_id = ?", service.Model.ID).Order("created_at desc").Find(&ports).Error; err != nil {
			log.Info(err.Error())
		}
		service.Ports = ports

		// Get deploy strategy
		deployStrategy := model.ServiceDeploymentStrategy{}
		if err := r.DB.Where("service_id = ?", service.Model.ID).First(&deployStrategy).Error; err != nil {
			log.Info(err.Error())
		}
		service.DeploymentStrategy = deployStrategy

		// Get readiness probe
		readinessProbe := model.ServiceHealthProbe{}
		rpHeaders := []model.ServiceHealthProbeHttpHeader{}
		if err := r.DB.Where("service_id = ? and type = ?", service.Model.ID, string(plugins.GetType("readinessProbe"))).First(&readinessProbe).Error; err != nil {
			log.Info(err.Error())
		}

		if err := r.DB.Where("health_probe_id = ?", readinessProbe.Model.ID).Find(&rpHeaders).Error; err != nil {
			log.Info(err.Error())
		}

		readinessProbe.HttpHeaders = rpHeaders
		service.ReadinessProbe = readinessProbe

		// get liveness probe
		livenessProbe := model.ServiceHealthProbe{}
		lpHeaders := []model.ServiceHealthProbeHttpHeader{}
		if err := r.DB.Where("service_id = ? and type = ?", service.Model.ID, string(plugins.GetType("livenessProbe"))).First(&livenessProbe).Error; err != nil {
			log.Info(err.Error())
		}

		if err := r.DB.Where("health_probe_id = ?", livenessProbe.Model.ID).Find(&lpHeaders).Error; err != nil {
			log.Info(err.Error())
		}

		livenessProbe.HttpHeaders = lpHeaders
		service.LivenessProbe = livenessProbe

		outServices = append(outServices, service)
	}

	out, err := yaml.Marshal(outServices)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
