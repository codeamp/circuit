package helpers

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/constants"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

func CreateServiceInDB(tx *gorm.DB, serviceInput *model.ServiceInput) (*model.Service, error) {
	// Check service name length
	if len(serviceInput.Name) > 63 {
		return nil, fmt.Errorf("Service name cannot be longer than 63 characters.")
	}

	// Check if project can create service in environment
	if tx.Where("environment_id = ? and project_id = ?", serviceInput.EnvironmentID, serviceInput.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
		return nil, fmt.Errorf("Project not allowed to create service in given environment")
	}

	// Check if service name already exists in environment and project
	if err := tx.Where("environment_id = ? and project_id = ? and name = ?",
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
	if err := tx.Where("is_default = ?", true).First(&defaultServiceSpec).Error; err != nil {
		return nil, fmt.Errorf("no default service spec found")
	}

	var deploymentStrategy model.ServiceDeploymentStrategy
	if serviceInput.DeploymentStrategy != nil {
		deploymentStrategy, err = ValidateDeploymentStrategyInput(serviceInput.DeploymentStrategy)
		if err != nil {
			return nil, err
		}
	}

	var livenessProbe model.ServiceHealthProbe
	if serviceInput.LivenessProbe != nil {
		probeType := plugins.GetType("livenessProbe")
		probe := serviceInput.LivenessProbe
		probe.Type = &probeType
		livenessProbe, err = ValidateHealthProbe(*probe)
		if err != nil {
			return nil, err
		}
	}

	var readinessProbe model.ServiceHealthProbe
	if serviceInput.ReadinessProbe != nil {
		probeType := plugins.GetType("readinessProbe")
		probe := serviceInput.ReadinessProbe
		probe.Type = &probeType
		readinessProbe, err = ValidateHealthProbe(*probe)
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

	if serviceInput.Ports != nil && len(*serviceInput.Ports) > 0 {
		if serviceInput.Type == string(plugins.GetType("general")) {
			for _, cp := range *serviceInput.Ports {
				servicePort := model.ServicePort{
					ServiceID: service.ID,
					Port:      cp.Port,
					Protocol:  cp.Protocol,
				}
				tx.Create(&servicePort)
			}
		} else {
			return nil, fmt.Errorf("Can only create ports if the service type is general")
		}
	}

	return &service, nil
}

func ValidateHealthProbe(input model.ServiceHealthProbeInput) (model.ServiceHealthProbe, error) {
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

func ValidateDeploymentStrategyInput(input *model.DeploymentStrategyInput) (model.ServiceDeploymentStrategy, error) {
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
