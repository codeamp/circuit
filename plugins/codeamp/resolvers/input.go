package codeamp_resolvers

// ServicePortInput
type ServicePortInput struct {
	// Port
	Port string `json:"port"`
	// Protocol
	Protocol string `json:"protocol"`
}

// EnvironmentInput
type EnvironmentInput struct {
	// ID
	ID *string `json:"id"`
	// Name
	Name string `json:"name"`
	// Color
	Color string `json:"color"`
}

// EnvironmentVariableInput
type EnvironmentVariableInput struct {
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
	// ProjectId
	ProjectId *string `json:"projectId"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
	// IsSecret
	IsSecret bool `json:"isSecret"`
}

// ExtensionInput
type ExtensionInput struct {
	// ID
	ID *string `json:"id"`
	// ProjectId
	ProjectId string `json:"projectId"`
	// ExtensionSpecId
	ExtensionSpecId string `json:"extensionSpecId"`
	// Config
	Config *JSON `json:"config"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
}

// ExtensionSpecInput
type ExtensionSpecInput struct {
	// ID
	ID *string `json:"id"`
	// Name
	Name string `json:"name"`
	// Key
	Key string `json:"key"`
	// Component
	Component string `json:"component"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
	// Config
	Config *JSON `json:"config"`
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
	// Bookmarked
	Bookmarked *bool `json:"bookmarked"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
}

// ReleaseInput
type ReleaseInput struct {
	// ID
	ID *string `json:"id"`
	// HeadFeatureId
	HeadFeatureId string `json:"headFeatureId"`
	// ProjectId
	ProjectId string `json:"projectId"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
}

// ServiceInput
type ServiceInput struct {
	// ID
	ID *string `json:"id"`
	// ProjectId
	ProjectId string `json:"projectId"`
	// Command
	Command string `json:"command"`
	// Name
	Name string `json:"name"`
	// ServiceSpecId
	ServiceSpecId string `json:"serviceSpecId"`
	// Count
	Count string `json:"count"`
	// ContainerPorts
	Ports *[]ServicePortInput `json:"ports"`
	// Type
	Type string `json:"type"`
	// EnvironmentId
	EnvironmentId string `json:"environmentId"`
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
}
