package model

import "github.com/codeamp/circuit/plugins"

// ServicePortInput
type ServicePortInput struct {
	// Port
	Port int32 `json:"port,string"`
	// Protocol
	Protocol string `json:"protocol"`
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
	Command string `json:"command"`
	// Name
	Name string `json:"name"`
	// Count
	Count int32 `json:"count,string"`
	// ContainerPorts
	Ports *[]ServicePortInput `json:"ports"`
	// Type
	Type string `json:"type"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
	// DeploymentStrategy
	DeploymentStrategy *DeploymentStrategyInput `json:"deploymentStrategy"`
	// ReadinessProbe
	ReadinessProbe *ServiceHealthProbeInput `json:"readinessProbe"`
	// LivenessProbe
	LivenessProbe *ServiceHealthProbeInput `json:"livenessProbe"`
	// PreStopHook
	PreStopHook *string `json"preStopHook"`
}

type DeploymentStrategyInput struct {
	// Type
	Type plugins.Type `json:"type"`
	// MaxUnavailable
	MaxUnavailable int32 `json:"maxUnavailable,string"`
	// MaxSurge
	MaxSurge int32 `json:"maxSurge,string"`
}

// ServiceHealthProbe is used for readiness/liveness health checks for services
// Further documentation can be found here: https://kubernetes.io/docs/tasks/configure-pod-container/configure-liveness-readiness-probes/
type ServiceHealthProbeInput struct {
	// Type currently supports ReadinessProbe and LivenessProbe
	Type *plugins.Type `json:"type"`
	//Method supports `exec`, `http`, and `tcp`
	Method string `json:"method"`
	// Command is only evaluated if Method is `exec`
	Command *string `json:"command"`
	// Port is only evaluated if Method is either `http` or `tcp`
	Port *int32 `json:"port,string"`
	// Scheme accepts `http` or `https` - it is only evaluated if Method is `http`
	Scheme *string `json:"scheme"`
	// Path is only evaluated if Method is `http`
	Path *string `json:"path"`
	// InitialDelaySeconds is the delay before the probe begins to evaluate service health
	InitialDelaySeconds *int32 `json:"initialDelaySeconds,string"`
	// PeriodSeconds is how frequently the probe is executed
	PeriodSeconds *int32 `json:"periodSeconds,string"`
	// TimeoutSeconds is the number of seconds before the probe times out
	TimeoutSeconds *int32 `json:"timeoutSeconds,string"`
	// SuccessThreshold minimum consecutive success before the probe is considered successfull
	SuccessThreshold *int32 `json:"successThreshold,string"`
	// FailureThreshold is the number of attempts before a probe is considered failed
	FailureThreshold *int32 `json:"failureThreshold,string"`
	// HealthProbeHttpHeaders
	HttpHeaders *[]HealthProbeHttpHeaderInput `json:"httpHeaders"`
}

type HealthProbeHttpHeaderInput struct {
	Name  string `json:"name"`
	Value string `json:"value"`
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
