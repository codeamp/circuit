package model

import "github.com/codeamp/circuit/plugins"

// ServicePortInput
type ServicePortInput struct {
	// Port
	Port int32 `json:"port,string" yaml:"port"`
	// Protocol
	Protocol string `json:"protocol" yaml:"protocol"`
}

// EnvironmentInput
type EnvironmentInput struct {
	// ID
	ID *string `json:"id"`
	// Name
	Name string `json:"name"`
	// Key
	Key string `json:"key"`
	// IsDefault
	IsDefault bool `json:"isDefault"`
	// Color
	Color string `json:"color"`
}

// SecretInput
type SecretInput struct {
	// ID
	ID *string `json:"id"`
	// Key
	Key string `json:"key"`
	// Value
	Value string `json:"value"`
	// Type
	Type string `json:"type"`
	// Scope
	Scope string `json:"scope"`
	// ProjectID
	ProjectID *string `json:"projectID"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
	// IsSecret
	IsSecret bool `json:"isSecret"`
}

// ImportSecretsInput
type ImportSecretsInput struct {
	// Secrets YAML
	SecretsYAMLString string `json:"secretsYAMLString`
	ProjectID         string `json:"projectID"`
	UserID            string `json:"userID"`
	EnvironmentID     string `json:"environmentID"`
}

// ExportSecretsInput
type ExportSecretsInput struct {
	ProjectID     string `json:"projectID"`
	EnvironmentID string `json:"environmentID"`
}

// ProjectExtensionInput
type ProjectExtensionInput struct {
	// ID
	ID *string `json:"id"`
	// ProjectID
	ProjectID string `json:"projectID"`
	// ExtensionID
	ExtensionID string `json:"extID"`
	// Config
	Config JSON `json:"config"`
	// CustomConfig
	CustomConfig JSON `json:"customConfig"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
}

// ExtensionInput
type ExtensionInput struct {
	// ID
	ID *string `json:"id"`
	// Name
	Name string `json:"name"`
	// Key
	Key string `json:"key"`
	// Component
	Component string `json:"component"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
	// Cacheable
	Cacheable bool `json:"cacheable"`
	// Config
	Config JSON `json:"config"`
	// Type
	Type string `json:"type"`
}

// ProjectInput
type ProjectInput struct {
	// ID
	ID *string `json:"id"`
	// GitProtocol
	GitProtocol string `json:"gitProtocol"`
	// GitUrl
	GitUrl string `json:"gitUrl"`
	// GitBranch
	GitBranch *string `json:"gitBranch"`
	// Bookmarked
	Bookmarked *bool `json:"bookmarked"`
	// ContinuousDeploy
	ContinuousDeploy *bool `json:"continuousDeploy"`
	// EnvironmentID
	EnvironmentID *string `json:"environmentID"`
}

// ProjectSearchInput
type ProjectSearchInput struct {
	// Repository
	Repository *string `json:"repository"`
	// Bookmarked
	Bookmarked bool `json:"bookmarked"`
}

// PaginatorInput
type PaginatorInput struct {
	// Page
	Page *int32 `json:"page"`
	// Limit
	Limit *int32 `json:"limit"`
}

// ReleaseInput
type ReleaseInput struct {
	// ID
	ID *string `json:"id"`
	// HeadFeatureID
	HeadFeatureID string `json:"headFeatureID"`
	// ProjectID
	ProjectID string `json:"projectID"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
	// ForceRebuild
	ForceRebuild bool `json:"forceRebuild"`
}

// ServiceInput
type ServiceInput struct {
	// ID
	ID *string `json:"id"`
	// ProjectID
	ProjectID string `json:"projectID"`
	// Command
	Command string `json:"command" yaml:"command"`
	// Name
	Name string `json:"name" yaml:"name"`
	// Count
	Count int32 `json:"count,string" yaml:"count"`
	// ContainerPorts
	Ports *[]ServicePortInput `json:"ports" yaml:"ports"`
	// Type
	Type string `json:"type" yaml:"type"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
	// DeploymentStrategy
	DeploymentStrategy *DeploymentStrategyInput `json:"deploymentStrategy" yaml:"deploymentStrategy"`
	// ReadinessProbe
	ReadinessProbe *ServiceHealthProbeInput `json:"readinessProbe" yaml:"readinessProbe"`
	// LivenessProbe
	LivenessProbe *ServiceHealthProbeInput `json:"livenessProbe" yaml:"livenessProbe"`
	// PreStopHook
	PreStopHook *string `json"preStopHook" yaml:"preStopHook"`
}

// ImportServicesInput
type ImportServicesInput struct {
	ServicesYAMLString string `json:"servicesYAMLString"`
	ProjectID          string `json:"projectID"`
	EnvironmentID      string `json:"environmentID"`
}

// ImportServicesInput
type ImportServicesInput struct {
	ServicesYAMLString string `json:"servicesYAMLString"`
	ProjectID          string `json:"projectID"`
	EnvironmentID      string `json:"environmentID"`
}

type DeploymentStrategyInput struct {
	// Type
	Type plugins.Type `json:"type" yaml:"type"`
	// MaxUnavailable
	MaxUnavailable int32 `json:"maxUnavailable,string" yaml:"maxUnavailable"`
	// MaxSurge
	MaxSurge int32 `json:"maxSurge,string" yaml:"maxSurge"`
}

// ServiceHealthProbe is used for readiness/liveness health checks for services
// Further documentation can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
type ServiceHealthProbeInput struct {
	// Type currently supports ReadinessProbe and LivenessProbe
	Type *plugins.Type `json:"type" yaml:"type"`
	//Method supports `exec`, `http`, and `tcp`
	Method string `json:"method" yaml:"method"`
	// Command is only evaluated if Method is `exec`
	Command *string `json:"command" yaml:"command"`
	// Port is only evaluated if Method is either `http` or `tcp`
	Port *int32 `json:"port,string" yaml:"port"`
	// Scheme accepts `http` or `https` - it is only evaluated if Method is `http`
	Scheme *string `json:"scheme" yaml:"scheme"`
	// Path is only evaluated if Method is `http`
	Path *string `json:"path" yaml:"path"`
	// InitialDelaySeconds is the delay before the probe begins to evaluate service health
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,string" yaml:"initialDelaySeconds"`
	// PeriodSeconds is how frequently the probe is executed
	PeriodSeconds *int32 `json:"periodSeconds,string" yaml:"periodSeconds"`
	// TimeoutSeconds is the number of seconds before the probe times out
	TimeoutSeconds *int32 `json:"timeoutSeconds,string" yaml:"timeoutSeconds"`
	// SuccessThreshold minimum consecutive success before the probe is considered successfull
	SuccessThreshold *int32 `json:"successThreshold,string" yaml:"successThreshold"`
	// FailureThreshold is the number of attempts before a probe is considered failed
	FailureThreshold *int32 `json:"failureThreshold,string" yaml:"failureThreshold"`
	// HealthProbeHttpHeaders
	HttpHeaders *[]HealthProbeHttpHeaderInput `json:"httpHeaders" yaml:"httpHeaders"`
}

type HealthProbeHttpHeaderInput struct {
	Name  string `json:"name" yaml:"name"`
	Value string `json:"value" yaml:"value"`
}

// ServiceSpecInput
type ServiceSpecInput struct {
	// ID
	ID *string `json:"id"`
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
	// IsDefault
	IsDefault bool `json:"isDefault"`
}

// UserPermissionsInput
type UserPermissionsInput struct {
	UserID      string            `json:"userID"`
	Permissions []PermissionInput `json:"permissions"`
}

// PermissionInput
type PermissionInput struct {
	Value string `json:"value"`
	Grant bool   `json:"grant"`
}

// ProjectEnvironmentInput
type ProjectEnvironmentInput struct {
	EnvironmentID string `json:"environmentID"`
	Grant         bool   `json:"grant"`
}

// ProjectEnvironmentsInput
type ProjectEnvironmentsInput struct {
	ProjectID   string                    `json:"projectID"`
	Permissions []ProjectEnvironmentInput `json:"permissions"`
}
