package graphql_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// Service Resolver Mutation
type ServiceResolverMutation struct {
	DB *gorm.DB
}

func (r *ServiceResolverMutation) CreateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	// Check service name length
	if len(args.Service.Name) > 63 {
		return nil, fmt.Errorf("Service name cannot be longer than 63 characters.")
	}

	tx := r.DB.Begin()

	// Check if project can create service in environment
	if err := tx.Where("environment_id = ? and project_id = ?", args.Service.EnvironmentID, args.Service.ProjectID).Find(&model.ProjectEnvironment{}).Error; err != nil {
		log.ErrorWithFields(err.Error(), log.Fields{
			"environment_id": args.Service.EnvironmentID,
			"project_id":     args.Service.ProjectID,
		})
		return nil, fmt.Errorf("Project not allowed to create service in given environment")
	}

	projectID, err := uuid.FromString(args.Service.ProjectID)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	environmentID, err := uuid.FromString(args.Service.EnvironmentID)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	if args.Service.DeploymentStrategy != nil {
		deploymentStrategy, err = validateDeploymentStrategyInput(args.Service.DeploymentStrategy)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	var livenessProbe model.ServiceHealthProbe
	if args.Service.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := args.Service.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe model.ServiceHealthProbe
	if args.Service.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := args.Service.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = validateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var preStopHook string
	if args.Service.PreStopHook != nil {
		preStopHook = *args.Service.PreStopHook
	}

	service := model.Service{
		Name:               args.Service.Name,
		Command:            args.Service.Command,
		Type:               plugins.Type(args.Service.Type),
		Count:              args.Service.Count,
		ProjectID:          projectID,
		EnvironmentID:      environmentID,
		DeploymentStrategy: deploymentStrategy,
		LivenessProbe:      livenessProbe,
		ReadinessProbe:     readinessProbe,
		PreStopHook:        preStopHook,
	}

	tx.Create(&service)

	// Create service spec from default
	defaultServiceSpec := model.ServiceSpec{}
	if err := tx.Where("is_default= ?", true).First(&defaultServiceSpec).Error; err != nil {
		tx.Rollback()
		log.Info(err.Error())
		return nil, fmt.Errorf("no default service spec found")
	}

	serviceSpec := model.ServiceSpec{
		Name:                   "",
		CpuRequest:             defaultServiceSpec.CpuRequest,
		CpuLimit:               defaultServiceSpec.CpuLimit,
		MemoryRequest:          defaultServiceSpec.MemoryRequest,
		MemoryLimit:            defaultServiceSpec.MemoryLimit,
		TerminationGracePeriod: defaultServiceSpec.TerminationGracePeriod,
		ServiceID:              service.Model.ID,
		IsDefault:              false,
	}

	if err := tx.Create(&serviceSpec).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

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

	if args.Service.Ports != nil {
		for _, cp := range *args.Service.Ports {
			servicePort := model.ServicePort{
				ServiceID: service.ID,
				Port:      cp.Port,
				Protocol:  cp.Protocol,
			}
			if err := tx.Create(&servicePort).Error; err != nil {
				log.Error(err.Error())
				tx.Rollback()
				return nil, err
			}
		}
	}

	if err := tx.Commit().Error; err != nil {
		log.Error(err.Error())
		tx.Rollback()
		return nil, err
	}

	return &ServiceResolver{DBServiceResolver: &db_resolver.ServiceResolver{DB: r.DB, Service: service}}, nil
}

// UpdateService Update Service
func (r *ServiceResolverMutation) UpdateService(args *struct{ Service *model.ServiceInput }) (*ServiceResolver, error) {
	serviceID := uuid.FromStringOrNil(*args.Service.ID)

	if serviceID == uuid.Nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	var service model.Service
	if err := r.DB.Where("id = ?", serviceID).Find(&service).Error; err != nil {
		log.ErrorWithFields(err.Error(), log.Fields{
			"id": serviceID,
		})
		return nil, err
	}

	service.Command = args.Service.Command
	service.Name = args.Service.Name
	service.Type = plugins.Type(args.Service.Type)
	service.Count = args.Service.Count
	service.AutoscaleEnabled = args.Service.AutoscaleEnabled

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
