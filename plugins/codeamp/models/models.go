package models

import (
	"time"

	"github.com/codeamp/circuit/plugins"
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

type EnvironmentVariable struct {
	Model     `json:",inline"`
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	Type      string    `json:"type"`
	Version   int32     `json:"version"`
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	UserId    uuid.UUID `bson:"userId" json:"userId" gorm:"type:uuid"`
	Created   time.Time `json:"created"`
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

	Features []Feature
	Releases []Release
	Service  []Service
}

type Workflow struct {
	Model     `json:"inline"`
	Name      string    `json:"name"`
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
}

type ServiceSpec struct {
	Model                  `json:",inline"`
	Name                   string    `json:"name"`
	CpuRequest             string    `json:"cpuRequest"`
	CpuLimit               string    `json:"cpuLimit"`
	MemoryRequest          string    `json:"memoryRequest"`
	MemoryLimit            string    `json:"memoryLimit"`
	TerminationGracePeriod string    `json:"terminationGracePeriod"`
	Created                time.Time `json:"created"`
}

type ContainerPort struct {
	Model     `json:",inline"`
	ServiceId uuid.UUID `bson:"serviceId" json:"serviceId" gorm:"type:uuid"`
	Protocol  string    `json:"protocol"`
	Port      string    `json:"port"`
}

type Service struct {
	Model         `json:",inline"`
	ProjectId     uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	ServiceSpecId uuid.UUID `bson:"serviceSpecId" json:"serviceSpecId" gorm:"type:uuid"`
	Command       string    `json:"command"`
	Name          string    `json:"name"`
	OneShot       bool      `json:"oneShot"`
	Count         string    `json:"count"`
	Created       time.Time `json:"created"`

	Project        Project
	ServiceSpec    ServiceSpec
	ContainerPorts []ContainerPort
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

	Project Project
}

type Release struct {
	Model `json:",inline"`

	Project   Project
	ProjectId uuid.UUID `json:"projectId" gorm:"type:uuid"`

	User   User
	UserID uuid.UUID `json:"userId" gorm:"type:uuid"`

	HeadFeature   Feature
	HeadFeatureID uuid.UUID `json:"headFeatureId" gorm:"type:uuid"`

	TailFeature   Feature
	TailFeatureID uuid.UUID `json:"tailFeatureId" gorm:"type:uuid"`

	Secrets      []EnvironmentVariable `json:"secrets"`
	Services     []Service             `json:"services"`
	State        plugins.State         `json:"state"`
	StateMessage string                `json:"stateMessage"`
	Created      time.Time             `json:"created"`
}

type Bookmark struct {
	Model     `json:",inline"`
	UserId    uuid.UUID `json:"userId" gorm:"type:uuid"`
	ProjectId uuid.UUID `json:"projectId" gorm:"type:uuid"`
}

type ExtensionSpec struct {
	Model     `json:",inline"`
	Type      string          `json:"type"`
	Name      string          `json:"name"`
	Component string          `json:"component"`
	FormSpec  postgres.Hstore `json:"formSpec"`
	Created   time.Time       `json:"created"`
}

type Extension struct {
	Model           `json:",inline"`
	ProjectId       uuid.UUID       `json:"projectId" gorm:"type:uuid"`
	ExtensionSpecId uuid.UUID       `json:"extensionSpecId" gorm:"type:uuid"`
	Slug            string          `json:"slug"`
	State           plugins.State   `json:"state"`
	Artifacts       postgres.Hstore `json:"artifacts"`
	FormSpecValues  postgres.Hstore `json:"formSpecValues"`
}

type ReleaseExtension struct {
	Model             `json:",inline"`
	FeatureHash       string          `json:"featureHash"`
	ServicesSignature string          `json:"servicesSignature"` // services config snapshot
	SecretsSignature  string          `json:"secretsSignature"`  // build args + artifacts
	ExtensionId       uuid.UUID       `json:"extensionId" gorm:"type:uuid"`
	State             plugins.State   `json:"state"`
	StateMessage      string          `json:"stateMessage"`
	Artifacts         postgres.Hstore `json:"artifacts"` // captured on workflow success/ fail
}
