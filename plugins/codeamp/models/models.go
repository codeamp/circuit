package models

import (
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

type Model struct {
	ID        uuid.UUID  `sql:"type:uuid;default:uuid_generate_v4()" json:"id" gorm:"primary_key"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
	DeletedAt *time.Time `json:"deletedAt" sql:"index"`
}

type User struct {
	Model       `json:",inline"`
	Email       string `json:"email" gorm:"type:varchar(100);unique_index"`
	Password    string `json:"password" gorm:"type:varchar(255)"`
	Permissions []UserPermission
}

type UserPermission struct {
	Model  `json:",inline"`
	UserId uuid.UUID `json:"userId" gorm:"type:uuid"`
	Value  string    `json:"value"`
}

type Environment struct {
	Model     `json:",inline"`
	Name      string `json:"name"`
}

type EnvironmentVariableScope string

func GetEnvironmentVariableScope(s string) EnvironmentVariableScope {
	environmentVariableScopes := []string{
		"project",
		"extension",
		"global",
	}

	for _, environmentVariableScope := range environmentVariableScopes {
		if s == environmentVariableScope {
			return EnvironmentVariableScope(environmentVariableScope)
		}
	}

	log.Info(fmt.Sprintf("EnvironmentVariableScope not found: %s", s))

	return EnvironmentVariableScope("unknown")
}

type EnvironmentVariable struct {
	Model         `json:",inline"`
	Key           string                   `json:"key"`
	Value         EnvironmentVariableValue `json:"value"`
	Type          plugins.Type             `json:"type"`
	ProjectId     uuid.UUID                `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	Scope         EnvironmentVariableScope `json:"scope"`
	EnvironmentId uuid.UUID                `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
}

type EnvironmentVariableValue struct {
	Model                 `json:",inline"`
	EnvironmentVariableId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	Value                 string    `json:"value"`
	UserId                uuid.UUID `bson:"userId" json:"userId" gorm:"type:uuid"`
}

type Project struct {
	Model         `json:",inline"`
	Name          string `json:"name"`
	Slug          string `json:"slug"`
	Repository    string `json:"repository"`
	Secret        string `json:"-"`
	GitUrl        string `json:"GitUrl"`
	GitProtocol   string `json:"GitProtocol"`
	RsaPrivateKey string `json:"-"`
	RsaPublicKey  string `json:"rsaPublicKey"`
}

type Workflow struct {
	Model     `json:"inline"`
	Name      string    `json:"name"`
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
}

type ServiceSpec struct {
	Model                  `json:",inline"`
	Name                   string `json:"name"`
	CpuRequest             string `json:"cpuRequest"`
	CpuLimit               string `json:"cpuLimit"`
	MemoryRequest          string `json:"memoryRequest"`
	MemoryLimit            string `json:"memoryLimit"`
	TerminationGracePeriod string `json:"terminationGracePeriod"`
}

type ContainerPort struct {
	Model     `json:",inline"`
	ServiceId uuid.UUID `bson:"serviceId" json:"serviceId" gorm:"type:uuid"`
	Protocol  string    `json:"protocol"`
	Port      string    `json:"port"`
}

type Service struct {
	Model         `json:",inline"`
	ProjectId     uuid.UUID    `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	ServiceSpecId uuid.UUID    `bson:"serviceSpecId" json:"serviceSpecId" gorm:"type:uuid"`
	Command       string       `json:"command"`
	Name          string       `json:"name"`
	Type          plugins.Type `json:"type"`
	Count         string       `json:"count"`
	EnvironmentId uuid.UUID    `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
}

type Feature struct {
	Model      `json:",inline"`
	ProjectId  uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	Message    string    `json:"message"`
	User       string    `json:"user"`
	Hash       string    `json:"hash"`
	ParentHash string    `json:"parentHash"`
	Ref        string    `json:"ref"`
	Created    time.Time `json:"created"`
}

type GitBranch struct {
	Model     `json:"inline"`
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	Name      string    `json:"name"`
}

type Release struct {
	Model         `json:",inline"`
	State         plugins.State `json:"state"`
	StateMessage  string        `json:"stateMessage"`
	Project       Project
	ProjectId     uuid.UUID `json:"projectId" gorm:"type:uuid"`
	User          User
	UserID        uuid.UUID `json:"userId" gorm:"type:uuid"`
	HeadFeature   Feature
	HeadFeatureID uuid.UUID `json:"headFeatureId" gorm:"type:uuid"`
	TailFeature   Feature
	TailFeatureID uuid.UUID             `json:"tailFeatureId" gorm:"type:uuid"`
	Secrets       []EnvironmentVariable `json:"secrets"`
	Services      []Service             `json:"services"`
	Artifacts     postgres.Jsonb        `json:"artifacts" gorm:"type:jsonb;not null"`
	Finished      time.Time
	EnvironmentId uuid.UUID `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
}

type Bookmark struct {
	Model     `json:",inline"`
	UserId    uuid.UUID `json:"userId" gorm:"type:uuid"`
	ProjectId uuid.UUID `json:"projectId" gorm:"type:uuid"`
}

type ExtensionSpec struct {
	Model         `json:",inline"`
	Type          plugins.Type   `json:"type"`
	Key           string         `json:"key"`
	Name          string         `json:"name"`
	Component     string         `json:"component"`
	EnvironmentId uuid.UUID      `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
	Config        postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
}

type Extension struct {
	Model           `json:",inline"`
	ProjectId       uuid.UUID      `json:"projectId" gorm:"type:uuid"`
	ExtensionSpecId uuid.UUID      `json:"extensionSpecId" gorm:"type:uuid"`
	State           plugins.State  `json:"state"`
	Artifacts       postgres.Jsonb `json:"artifacts" gorm:"type:jsonb;not null"`
	Config          postgres.Jsonb `json:"config" gorm:"type:jsonb;not null"`
	EnvironmentId   uuid.UUID      `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
}

type ReleaseExtension struct {
	Model             `json:",inline"`
	ReleaseId         uuid.UUID      `json:"releaseId" gorm:"type:uuid"`
	FeatureHash       string         `json:"featureHash"`
	ServicesSignature string         `json:"servicesSignature"` // services config snapshot
	SecretsSignature  string         `json:"secretsSignature"`  // build args + artifacts
	ExtensionId       uuid.UUID      `json:"extensionId" gorm:"type:uuid"`
	State             plugins.State  `json:"state"`
	StateMessage      string         `json:"stateMessage"`
	Type              plugins.Type   `json:"type"`
	Artifacts         postgres.Jsonb `json:"artifacts" gorm:"type:jsonb;not null"` // captured on workflow success/ fail
	Finished          time.Time
}

type ReleaseDeployment struct {
	Model        `json:",inline"`
	ReleaseId    uuid.UUID      `json:"releaseId" gorm:"type:uuid"`
	State        plugins.State  `json:"state"`
	StateMessage string         `json:"stateMessage"`
	Artifacts    postgres.Jsonb `json:"artifacts" gorm:"type:jsonb;not null"` // captured on workflow success/ fail
	Finished     time.Time
}
