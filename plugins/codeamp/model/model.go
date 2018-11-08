package model

import (
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

// Default fields for a model
type Model struct {
	// ID
	ID uuid.UUID `sql:"type:uuid;default:uuid_generate_v4()" json:"id" gorm:"primary_key"`
	// CreatedAt
	CreatedAt time.Time `json:"createdAt"`
	// UpdatedAt
	UpdatedAt time.Time `json:"updatedAt"`
	// DeletedAt
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

// User
type User struct {
	Model `json:",inline"`
	// Email
	Email string `json:"email"`
	// Password
	Password string `json:"password" gorm:"type:varchar(255)"`
	// Permissions
	Permissions []UserPermission `json:"permissions"`
}

// User permission
type UserPermission struct {
	Model `json:",inline"`
	// UserID
	UserID uuid.UUID `json:"userID" gorm:"type:uuid"`
	// Value
	Value string `json:"value"`
}

// Claims
type Claims struct {
	UserID      string   `json:"userID"`
	Email       string   `json:"email"`
	Verified    bool     `json:"email_verified"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	TokenError  string   `json:"tokenError"`
}

// Release
type Release struct {
	Model `json:",inline"`
	// State
	State transistor.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// User
	User User
	// UserID
	UserID uuid.UUID `json:"userID" gorm:"type:uuid"`
	// HeadFeatureID
	HeadFeatureID uuid.UUID `json:"headFeatureID" gorm:"type:uuid"`
	// TailFeatureID
	TailFeatureID uuid.UUID `json:"tailFeatureID" gorm:"type:uuid"`
	// Services
	Services postgres.Jsonb `json:"services" gorm:"type:jsonb;"`
	// Secrets
	Secrets postgres.Jsonb `json:"secrets" gorm:"type:jsonb;"`
	// ProjectExtensions
	ProjectExtensions postgres.Jsonb `json:"extensions" gorm:"type:jsonb;"`
	// EnvironmentID
	EnvironmentID uuid.UUID `json:"environmentID" gorm:"type:uuid"`
	// Started
	Started time.Time
	// Finished
	Finished time.Time
	// ForceRebuild
	ForceRebuild bool `json:"forceRebuild"`
	IsRollback   bool `json:"isRollback"`
}

// ServiceSpec
type ServiceSpec struct {
	Model `json:",inline"`
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
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"serviceID" gorm:"type:uuid"`
}

// Service
type Service struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// ServiceSpecID
	ServiceSpecID uuid.UUID `bson:"serviceSpecID" json:"serviceSpecID" gorm:"type:uuid"`
	// Command
	Command string `json:"command"`
	// Name
	Name string `json:"name"`
	// Type
	Type plugins.Type `json:"type"`
	// Count
	Count int32 `json:"count,string"`
	// Ports
	Ports []ServicePort `json:"servicePorts"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
	// DeploymentStrategy
	DeploymentStrategy ServiceDeploymentStrategy `json:"deploymentStrategy"`
	// ReadinessProbe
	ReadinessProbe ServiceHealthProbe `json:"readinessProbe"`
	// LivenessProbe
	LivenessProbe ServiceHealthProbe `json:"livenessProbe"`
	// PreStopHook
	PreStopHook string `json:"preStopHook"`
}

// ServicePort
type ServicePort struct {
	Model `json:",inline"`
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"serviceID" gorm:"type:uuid"`
	// Protocol
	Protocol string `json:"protocol"`
	// Port
	Port int32 `json:"port"`
}

// DeploymentStrategy
type ServiceDeploymentStrategy struct {
	// Model
	Model `json:",inline"`
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"serviceID" gorm:"type:uuid"`
	// Type
	Type plugins.Type `json:"type"`
	// MaxUnavailable
	MaxUnavailable int32 `json:"maxUnavailable"`
	// MaxSurge
	MaxSurge int32 `json:"maxSurge"`
}

// ServiceHealthProbe
type ServiceHealthProbe struct {
	// Model
	Model `json:",inline"`
	// ServiceID
	ServiceID uuid.UUID `bson:"serviceID" json:"serviceID" gorm:"type:uuid"`
	// Type: required; accepts `readinessProbe` and `livenessProbe`
	Type plugins.Type `json:"type"`
	// Method: required; accepts `exec`, `http`, and `tcp`
	Method string `json:"method"`
	// Command: Required with Method `exec`
	Command string `json:"command"`
	// Port: Required with Method `http` or `tcp`
	Port int32 `json:"port"`
	// Scheme: required with method `http`; accepts `http` or `https`
	Scheme string `json:"scheme"`
	// Path: required with Method `http`
	Path string `json:"path"`
	//HTTPHeaders is a slice of HTTP headers used by http probes
	HttpHeaders []ServiceHealthProbeHttpHeader `json:"httpHeaders"`
	// InitialDelaySeconds is the delay before the probe begins to evaluate service health
	InitialDelaySeconds int32 `json:"initialDelaySeconds"`
	// PeriodSeconds is how frequently the probe is executed
	PeriodSeconds int32 `json:"periodSeconds"`
	// TimeoutSeconds is the number of seconds before the probe times out
	TimeoutSeconds int32 `json:"timeoutSeconds"`
	// SuccessThreshold minimum consecutive success before the probe is considered successfull
	SuccessThreshold int32 `json:"successThreshold"`
	// FailureThreshold is the number of attempts before a probe is considered failed
	FailureThreshold int32 `json:"failureThreshold"`
}

type ServiceHealthProbeHttpHeader struct {
	Model         `json:",inline"`
	HealthProbeID uuid.UUID `bson:"healthProbeID" json:"healthProbeID" gorm:"type:uuid"`
	Name          string    `json:"name"`
	Value         string    `json:"value"`
}

// Secret
type Secret struct {
	Model `json:",inline"`
	// Key
	Key string `json:"key"`
	// Value
	Value SecretValue `json:"value"`
	// Type
	Type plugins.Type `json:"type"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// Scope
	Scope SecretScope `json:"scope"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
	// IsSecret
	IsSecret bool `json:"isSecret"`
}

type SecretValue struct {
	Model `json:",inline"`
	// SecretID
	SecretID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// Value
	Value string `json:"value"`
	// UserID
	UserID uuid.UUID `bson:"userID" json:"userID" gorm:"type:uuid"`
}

type SecretScope string

// ReleaseExtension
type ReleaseExtension struct {
	Model `json:",inline"`
	// ReleaseID
	ReleaseID uuid.UUID `json:"releaseID" gorm:"type:uuid"`
	// FetureHash
	FeatureHash string `json:"featureHash"`
	// ServicesSignature
	ServicesSignature string `json:"servicesSignature"`
	// SecretsSignature
	SecretsSignature string `json:"secretsSignature"`
	// ProjectExtensionID
	ProjectExtensionID uuid.UUID `json:"extensionID" gorm:"type:uuid"`
	// State
	State transistor.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// Type
	Type plugins.Type `json:"type"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"` // captured on workflow success/ fail
	// Finished
	Finished time.Time
	// Started
	Started time.Time
}

// Project
type Project struct {
	Model `json:",inline"`
	// Name
	Name string `json:"name"`
	// Slug
	Slug string `json:"slug"`
	// Repository
	Repository string `json:"repository"`
	// Secret
	Secret string `json:"-"`
	// GitUrl
	GitUrl string `json:"GitUrl"`
	// GitProtocol
	GitProtocol string `json:"GitProtocol"`
	// RsaPrivateKey
	RsaPrivateKey string `json:"-"`
	// RsaPublicKey
	RsaPublicKey string `json:"rsaPublicKey"`
}

// Project settings
type ProjectSettings struct {
	Model `json:"inline"`
	// EnvironmentID
	EnvironmentID uuid.UUID `json:"environmentID" gorm:"type:uuid"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// GitBranch
	GitBranch string `json:"gitBranch"`
	//ContinuousDeploy
	ContinuousDeploy bool `json:"continuousDeploy"`
}

// ProjectEnvironment
type ProjectEnvironment struct {
	Model `json:"inline"`
	// EnvironmentID
	EnvironmentID uuid.UUID `json:"environmentID" gorm:"type:uuid"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
}

// ProjectEnvironment
type ProjectBookmark struct {
	Model `json:"inline"`
	// UserID
	UserID uuid.UUID `json:"userID" gorm:"type:uuid"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
}

// ProjectExtension
type ProjectExtension struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `json:"projectID" gorm:"type:uuid"`
	// ExtensionID
	ExtensionID uuid.UUID `json:"extID" gorm:"type:uuid"`
	// State
	State transistor.State `json:"state"`
	// StateMessage
	StateMessage string `json:"stateMessage"`
	// Artifacts
	Artifacts postgres.Jsonb `json:"artifacts" gorm:"type:jsonb"`
	// Config
	Config postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
	// CustomConfig
	CustomConfig postgres.Jsonb `json:"customConfig" gorm:"type:jsonb;not null"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
}

// Environment Environment
type Environment struct {
	Model `json:",inline"`
	// Name
	Name string `json:"name"`
	// Key
	Key string `json:"key"`
	// Is Default
	IsDefault bool `json:"isDefault"`
	// Color
	Color string `json:"color"`
}

// Extension spec
type Extension struct {
	Model `json:",inline"`
	// Type
	Type plugins.Type `json:"type"`
	// Key
	Key string `json:"key"`
	// Name
	Name string `json:"name"`
	// Component
	Component string `json:"component"`
	// Cacheable
	Cacheable bool `json:"cacheable"`
	// EnvironmentID
	EnvironmentID uuid.UUID `bson:"environmentID" json:"environmentID" gorm:"type:uuid"`
	// Config
	Config postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
}

// Feature
type Feature struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// Message
	Message string `json:"message"`
	// User
	User string `json:"user"`
	// Hash
	Hash string `json:"hash"`
	// ParentHash
	ParentHash string `json:"parentHash"`
	// Ref
	Ref string `json:"ref"`
	// Created
	Created time.Time `json:"created"`
}

// ExtConfig
type ExtConfig struct {
	Key           string `json:"key"`
	Value         string `json:"value"`
	Secret        bool   `json:"secret"`
	AllowOverride bool   `json:"allowOverride"`
}

/////////////////////////////
/////////////////////////////
func (s *Secret) AfterFind(tx *gorm.DB) (err error) {
	if s.Value == (SecretValue{}) {
		var secretValue SecretValue
		tx.Where("secret_id = ?", s.Model.ID).Order("created_at desc").FirstOrInit(&secretValue)
		s.Value = secretValue
	}
	return
}
