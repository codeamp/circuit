package codeamp_resolvers

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// EnvironmentVariable
type EnvironmentVariable struct {
	Model `json:",inline"`
	// Key
	Key string `json:"key"`
	// Value
	Value EnvironmentVariableValue `json:"value"`
	// Type
	Type plugins.Type `json:"type"`
	// ProjectId
	ProjectId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	// Scope
	Scope EnvironmentVariableScope `json:"scope"`
	// EnvironmentId
	EnvironmentId uuid.UUID `bson:"environmentId" json:"environmentId" gorm:"type:uuid"`
	// IsSecret
	IsSecret bool `json:"isSecret"`
}

type EnvironmentVariableValue struct {
	Model `json:",inline"`
	// EnvironmentVariableId
	EnvironmentVariableId uuid.UUID `bson:"projectId" json:"projectId" gorm:"type:uuid"`
	// Value
	Value string `json:"value"`
	// UserId
	UserId uuid.UUID `bson:"userId" json:"userId" gorm:"type:uuid"`
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

// EnvironmentVariableResolver resolver for EnvironmentVariable
type EnvironmentVariableResolver struct {
	EnvironmentVariable
	EnvironmentVariableValue EnvironmentVariableValue
	DB                       *gorm.DB
}

// ID
func (r *EnvironmentVariableResolver) ID() graphql.ID {
	return graphql.ID(r.EnvironmentVariable.Model.ID.String())
}

// Key
func (r *EnvironmentVariableResolver) Key() string {
	return r.EnvironmentVariable.Key
}

// Value
func (r *EnvironmentVariableResolver) Value() string {
	return r.EnvironmentVariableValue.Value
}

// Scope
func (r *EnvironmentVariableResolver) Scope() string {
	return string(r.EnvironmentVariable.Scope)
}

// Project
func (r *EnvironmentVariableResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.EnvironmentVariable).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// User
func (r *EnvironmentVariableResolver) User() *UserResolver {
	var user User

	r.DB.Model(r.EnvironmentVariableValue).Related(&user)

	return &UserResolver{DB: r.DB, User: user}
}

// Type
func (r *EnvironmentVariableResolver) Type() string {
	return string(r.EnvironmentVariable.Type)
}

// Versions
func (r *EnvironmentVariableResolver) Versions() ([]*EnvironmentVariableResolver, error) {
	var envVarValues []EnvironmentVariableValue
	var envVarResolvers []*EnvironmentVariableResolver

	r.DB.Where("environment_variable_id = ?", r.EnvironmentVariable.Model.ID).Order("created_at desc").Find(&envVarValues)

	for _, envVarValue := range envVarValues {
		envVarResolvers = append(envVarResolvers, &EnvironmentVariableResolver{DB: r.DB, EnvironmentVariable: r.EnvironmentVariable, EnvironmentVariableValue: envVarValue})
	}

	return envVarResolvers, nil
}

// Environment
func (r *EnvironmentVariableResolver) Environment() *EnvironmentResolver {
	var env Environment

	r.DB.Model(r.EnvironmentVariable).Related(&env)

	return &EnvironmentResolver{DB: r.DB, Environment: env}
}

// Created
func (r *EnvironmentVariableResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentVariable.Model.CreatedAt}
}

// IsSecret
func (r *EnvironmentVariableResolver) IsSecret() bool {
	return r.EnvironmentVariable.IsSecret
}

func (r *EnvironmentVariableResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.EnvironmentVariable)
}

func (r *EnvironmentVariableResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.EnvironmentVariable)
}
