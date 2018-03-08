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
	// Key
	Key string `json:"key"`
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
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
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
	// ServiceSpecID
	ServiceSpecID string `json:"serviceSpecID"`
	// Count
	Count string `json:"count"`
	// ContainerPorts
	Ports *[]ServicePortInput `json:"ports"`
	// Type
	Type string `json:"type"`
	// EnvironmentID
	EnvironmentID string `json:"environmentID"`
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

// ProjectPermissionInput
type ProjectPermissionInput struct {
	EnvironmentID string `json:"environmentID"`
	Grant         bool   `json:"grant"`
}

// ProjectPermissionsInput
type ProjectPermissionsInput struct {
	ProjectID   string                   `json:"projectID"`
	Permissions []ProjectPermissionInput `json:"permissions"`
}
