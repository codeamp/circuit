package graphql_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/constants"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
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

	service, err := r.createServiceInDB(tx, args.Service)
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
		livenessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe = model.ServiceHealthProbe{}
	if args.Service.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := args.Service.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = validateHealthProbe(*probe)
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
	updatedDeploymentStrategy, err := validateDeploymentStrategyInput(args.Service.DeploymentStrategy)
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

// DeleteService Delete service
func (r *ServiceResolverMutation) DeleteService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID, err := uuid.FromString(*args.Service.ID)

	if err != nil {
		return nil, err
	}

	var service model.Service

	r.DB.Where("id = ?", serviceID).Find(&service)
	r.DB.Delete(&service)

	// delete all previous container ports
	var servicePorts []model.ServicePort
	r.DB.Where("service_id = ?", serviceID).Find(&servicePorts)

	// delete all container ports
	for _, cp := range servicePorts {
		r.DB.Delete(&cp)
	}

	var healthProbes []model.ServiceHealthProbe
	r.DB.Where("service_id = ?", serviceID).Find(&healthProbes)
	for _, probe := range healthProbes {
		var headers []model.ServiceHealthProbeHttpHeader
		r.DB.Where("health_probe_id = ?", probe.ID).Find(&headers)
		for _, header := range headers {
			r.DB.Delete(&header)
		}
		r.DB.Delete(&probe)
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	r.DB.Where("service_id = ?", serviceID).Find(&deploymentStrategy)
	r.DB.Delete(&deploymentStrategy)

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

func validateHealthProbe(input model.ServiceHealthProbeInput) (model.ServiceHealthProbe, error) {
	healthProbe := model.ServiceHealthProbe{}

	switch probeType := *input.Type; probeType {
	case plugins.GetType("livenessProbe"), plugins.GetType("readinessProbe"):
		healthProbe.Type = probeType
		if input.InitialDelaySeconds != nil {
			healthProbe.InitialDelaySeconds = *input.InitialDelaySeconds
		}
		if input.PeriodSeconds != nil {
			healthProbe.PeriodSeconds = *input.PeriodSeconds
		}
		if input.TimeoutSeconds != nil {
			healthProbe.TimeoutSeconds = *input.TimeoutSeconds
		}
		if input.SuccessThreshold != nil {
			healthProbe.SuccessThreshold = *input.SuccessThreshold
		}
		if input.FailureThreshold != nil {
			healthProbe.FailureThreshold = *input.FailureThreshold
		}
	default:
		return model.ServiceHealthProbe{}, fmt.Errorf("Unsuported Probe Type %s", string(*input.Type))
	}

	switch probeMethod := input.Method; probeMethod {
	case "default", "":
		return model.ServiceHealthProbe{}, nil
	case "exec":
		healthProbe.Method = input.Method
		if input.Command == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("Command is required if Probe method is exec")
		}
		healthProbe.Command = *input.Command
	case "http":
		healthProbe.Method = input.Method
		if input.Port == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe require a port to be set")
		}
		healthProbe.Port = *input.Port
		if input.Path == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe requires a path to be set")
		}
		healthProbe.Path = *input.Path

		// httpStr := "http"
		// httpsStr := "https"
		if input.Scheme == nil || (*input.Scheme != "http" && *input.Scheme != "https") {
			return model.ServiceHealthProbe{}, fmt.Errorf("http probe requires scheme to be set to either http or https")
		}
		healthProbe.Scheme = *input.Scheme
	case "tcp":
		healthProbe.Method = input.Method
		if input.Port == nil {
			return model.ServiceHealthProbe{}, fmt.Errorf("tcp probe requires a port to be set")
		}
		healthProbe.Port = *input.Port
	default:
		return model.ServiceHealthProbe{}, fmt.Errorf("Unsuported Probe Method %s", string(input.Method))
	}

	if input.HttpHeaders != nil {
		for _, headerInput := range *input.HttpHeaders {
			header := model.ServiceHealthProbeHttpHeader{
				Name:          headerInput.Name,
				Value:         headerInput.Value,
				HealthProbeID: healthProbe.ID,
			}
			healthProbe.HttpHeaders = append(healthProbe.HttpHeaders, header)
		}

	}

	return healthProbe, nil
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

		service, err := r.createServiceInDB(tx, &service)
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

func validateDeploymentStrategyInput(input *model.DeploymentStrategyInput) (model.ServiceDeploymentStrategy, error) {
	switch strategy := input.Type; strategy {
	case plugins.GetType("default"), plugins.GetType("recreate"):
		return model.ServiceDeploymentStrategy{Type: plugins.Type(input.Type)}, nil
	case plugins.GetType("rollingUpdate"):
		if input.MaxUnavailable == 0 {
			return model.ServiceDeploymentStrategy{}, fmt.Errorf("RollingUpdate DeploymentStrategy requires a valid maxUnavailable parameter")
		}

		if input.MaxSurge == 0 {
			return model.ServiceDeploymentStrategy{}, fmt.Errorf("RollingUpdate DeploymentStrategy requires a valid maxSurge parameter")
		}
	default:
		return model.ServiceDeploymentStrategy{}, fmt.Errorf("Unsuported Deployment Strategy %s", input.Type)
	}

	deploymentStrategy := model.ServiceDeploymentStrategy{
		Type:           plugins.Type(input.Type),
		MaxUnavailable: input.MaxUnavailable,
		MaxSurge:       input.MaxSurge,
	}

	return deploymentStrategy, nil
}

func (r *ServiceResolverMutation) createServiceInDB(tx *gorm.DB, serviceInput *model.ServiceInput) (*model.Service, error) {
	// Check service name length
	if len(serviceInput.Name) > 63 {
		return nil, fmt.Errorf("Service name cannot be longer than 63 characters.")
	}

	// Check if project can create service in environment
	if r.DB.Where("environment_id = ? and project_id = ?", serviceInput.EnvironmentID, serviceInput.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
		return nil, fmt.Errorf("Project not allowed to create service in given environment")
	}

	// Check if service name already exists in environment and project
	if err := r.DB.Where("environment_id = ? and project_id = ? and name = ?",
		serviceInput.EnvironmentID, serviceInput.ProjectID, serviceInput.Name).First(&model.Service{}).Error; err == nil {
		return nil, fmt.Errorf(constants.ServiceAlreadyExistsErrMsg)
	}

	// Check if service type exists
	if string(plugins.GetType(serviceInput.Type)) == "unknown" {
		return nil, fmt.Errorf("Service type %s does not exist", serviceInput.Type)
	}

	projectID, err := uuid.FromString(serviceInput.ProjectID)
	if err != nil {
		return nil, err
	}

	environmentID, err := uuid.FromString(serviceInput.EnvironmentID)
	if err != nil {
		return nil, err
	}

	// Find the default service spec and create ServiceSpec specific for Service
	defaultServiceSpec := model.ServiceSpec{}
	if err := r.DB.Where("is_default = ?", true).First(&defaultServiceSpec).Error; err != nil {
		return nil, fmt.Errorf("no default service spec found")
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	if serviceInput.DeploymentStrategy != nil {
		deploymentStrategy, err = validateDeploymentStrategyInput(serviceInput.DeploymentStrategy)
		if err != nil {
			return nil, err
		}
	}

	var livenessProbe model.ServiceHealthProbe
	if serviceInput.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := serviceInput.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe model.ServiceHealthProbe
	if serviceInput.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := serviceInput.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var preStopHook string
	if serviceInput.PreStopHook != nil {
		preStopHook = *serviceInput.PreStopHook
	}

	service := model.Service{
		Name:               serviceInput.Name,
		Command:            serviceInput.Command,
		Type:               plugins.Type(serviceInput.Type),
		Count:              serviceInput.Count,
		ProjectID:          projectID,
		EnvironmentID:      environmentID,
		DeploymentStrategy: deploymentStrategy,
		LivenessProbe:      livenessProbe,
		ReadinessProbe:     readinessProbe,
		PreStopHook:        preStopHook,
	}

	tx.Create(&service)

	serviceSpec := model.ServiceSpec{
		Name:                   defaultServiceSpec.Name,
		CpuRequest:             defaultServiceSpec.CpuRequest,
		CpuLimit:               defaultServiceSpec.CpuLimit,
		MemoryRequest:          defaultServiceSpec.MemoryRequest,
		MemoryLimit:            defaultServiceSpec.MemoryLimit,
		TerminationGracePeriod: defaultServiceSpec.TerminationGracePeriod,
		ServiceID:              service.Model.ID,
		IsDefault:              false,
	}

	tx.Create(&serviceSpec)

	// Create Health Probe Headers
	if service.LivenessProbe.HttpHeaders != nil {
		for _, h := range service.LivenessProbe.HttpHeaders {
			h.HealthProbeID = service.LivenessProbe.ID
			tx.Create(&h)
		}
	}

	if service.ReadinessProbe.HttpHeaders != nil {
		for _, h := range service.ReadinessProbe.HttpHeaders {
			h.HealthProbeID = service.ReadinessProbe.ID
			tx.Create(&h)
		}
	}

	if serviceInput.Ports != nil && serviceInput.Type == string(plugins.GetType("general")) {
		for _, cp := range *serviceInput.Ports {
			servicePort := model.ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			tx.Create(&servicePort)
		}
	}

	return &service, nil
}
