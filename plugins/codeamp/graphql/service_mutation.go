package graphql_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/constants"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/helpers"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	yaml "gopkg.in/yaml.v2"
)

// Service Resolver Mutation
type ServiceResolverMutation struct {
	DB *gorm.DB
}

func (r *ServiceResolverMutation) CreateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	tx := r.DB.Begin()

	service, err := helpers.CreateServiceInDB(tx, args.Service)
	if err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return nil, err
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: *service}}, nil
}

// UpdateService Update Service
func (r *ServiceResolverMutation) UpdateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID := uuid.FromStringOrNil(*args.Service.ID)

	if serviceID == uuid.Nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	var service model.Service
	if r.DB.Where("id = ?", serviceID).Find(&service).RecordNotFound() {
		return nil, fmt.Errorf("Record not found with given argument id")
	}

	service.Command = args.Service.Command
	service.Name = args.Service.Name
	service.Type = plugins.Type(args.Service.Type)
	service.Count = args.Service.Count

	if err := r.DB.Save(&service).Error; err != nil {
		log.Error(err.Error())
		return nil, err
	}

	// delete all previous container ports
	var servicePorts []model.ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	// replace with current
	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := model.ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			r.DB.Create(&servicePort)
		}
	}

	var livenessProbe = model.ServiceHealthProbe{}
	var err error
	if args.Service.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := args.Service.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = helpers.ValidateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe = model.ServiceHealthProbe{}
	if args.Service.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := args.Service.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = helpers.ValidateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var oldHealthProbes []model.ServiceHealthProbe
	r.DB.Where("service_id = ?", serviceID).Find(&oldHealthProbes)
	for _, probe := range oldHealthProbes {
		var headers []model.ServiceHealthProbeHttpHeader
		r.DB.Where("health_probe_id = ?", probe.ID).Find(&headers)
		for _, header := range headers {
			r.DB.Delete(&header)
		}
		r.DB.Delete(&probe)
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	r.DB.Where("service_id = ?", serviceID).Find(&deploymentStrategy)
	updatedDeploymentStrategy, err := helpers.ValidateDeploymentStrategyInput(args.Service.DeploymentStrategy)
	if err != nil {
		return nil, err
	}

	deploymentStrategy.Type = updatedDeploymentStrategy.Type
	deploymentStrategy.MaxUnavailable = updatedDeploymentStrategy.MaxUnavailable
	deploymentStrategy.MaxSurge = updatedDeploymentStrategy.MaxSurge

	r.DB.Save(&deploymentStrategy)
	service.DeploymentStrategy = deploymentStrategy
	service.ReadinessProbe = readinessProbe
	service.LivenessProbe = livenessProbe

	var preStopHook string
	if args.Service.PreStopHook != nil {
		preStopHook = *args.Service.PreStopHook
	}
	service.PreStopHook = preStopHook
	r.DB.Save(&service)

	// Create Health Probe Headers
	for _, h := range service.LivenessProbe.HttpHeaders {
		h.HealthProbeID = service.LivenessProbe.ID
		r.DB.Create(&h)
	}

	for _, h := range service.ReadinessProbe.HttpHeaders {
		h.HealthProbeID = service.ReadinessProbe.ID
		r.DB.Create(&h)
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

// ImportServices takes in a YAML string of services as input
// and batch creates new services in the specified environment and project
func (r *ServiceResolverMutation) ImportServices(args *struct{ Services *model.ImportServicesInput }) ([]*ServiceResolver, error) {
	services := []model.ServiceInput{}
	serviceResolvers := []*ServiceResolver{}
	// unmarshal services yaml string into []model.Service{}
	err := yaml.Unmarshal([]byte(args.Services.ServicesYAMLString), &services)
	if err != nil {
		return nil, err
	}

	tx := r.DB.Begin()

	// for each one, call CreateService
	for _, service := range services {
		service.ProjectID = args.Services.ProjectID
		service.EnvironmentID = args.Services.EnvironmentID

		service, err := helpers.CreateServiceInDB(tx, &service)
		if err == nil {
			serviceResolvers = append(serviceResolvers, &ServiceResolver{
				DBServiceResolver: &db_resolver.ServiceResolver{
					Service: *service,
					DB:      r.DB,
				},
			})
			// check if it's just a service already created error. We can
			// continue creation if so
		} else if err != nil && err.Error() != constants.ServiceAlreadyExistsErrMsg {
			return []*ServiceResolver{}, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		return []*ServiceResolver{}, err
	}

	// return newly created services
	return serviceResolvers, nil
}

// DeleteService Delete service
func (r *ServiceResolverMutation) DeleteService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID, err := uuid.FromString(*args.Service.ID)

	if err != nil {
		return nil, err
	}

	var service model.Service

	if err := r.DB.Where("id = ?", serviceID).Find(&service).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Delete(&service).Error; err != nil {
		return nil, err
	}

	// delete all previous container ports
	var servicePorts []model.ServicePort
	if err := r.DB.Where("service_id = ?", serviceID).Find(&servicePorts).Error; err != nil {
		return nil, err
	}

	// delete all container ports
	for _, cp := range servicePorts {
		if err := r.DB.Delete(&cp).Error; err != nil {
			return nil, err
		}
	}

	var healthProbes []model.ServiceHealthProbe
	if err := r.DB.Where("service_id = ?", serviceID).Find(&healthProbes).Error; err != nil {
		return nil, err
	}

	for _, probe := range healthProbes {
		var headers []model.ServiceHealthProbeHttpHeader
		if err := r.DB.Where("health_probe_id = ?", probe.ID).Find(&headers).Error; err != nil {
			return nil, err
		}

		for _, header := range headers {
			if err := r.DB.Delete(&header).Error; err != nil {
				return nil, err
			}
		}

		if err := r.DB.Delete(&probe).Error; err != nil {
			return nil, err
		}
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	if err := r.DB.Where("service_id = ?", serviceID).Find(&deploymentStrategy).Error; err != nil {
		return nil, err
	}

	if err := r.DB.Delete(&deploymentStrategy).Error; err != nil {
		return nil, err
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}
