package model

import (
	"time"

	"github.com/codeamp/transistor"
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

//Claims
type Claims struct {
	UserID      string   `json:"userID"`
	Email       string   `json:"email"`
	Verified    bool     `json:"email_verified"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	TokenError  string   `json:"tokenError"`
}

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
	// FinishedAt
	FinishedAt time.Time
	// ForceRebuild
	ForceRebuild bool `json:"forceRebuild"`
}
