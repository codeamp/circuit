package codeamp_resolvers

import (
	"time"

	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
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

// Resolver is the main resolver for all queries
type Resolver struct {
	// DB
	DB *gorm.DB
	// Events
	Events chan transistor.Event
}
