package graphql_resolver

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

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

func (s *Secret) AfterFind(tx *gorm.DB) (err error) {
	if s.Value == (SecretValue{}) {
		var secretValue SecretValue
		tx.Where("secret_id = ?", s.Model.ID).Order("created_at desc").FirstOrInit(&secretValue)
		s.Value = secretValue
	}
	return
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

func GetSecretScope(s string) SecretScope {
	secretScopes := []string{
		"project",
		"extension",
		"global",
	}

	for _, secretScope := range secretScopes {
		if s == secretScope {
			return SecretScope(secretScope)
		}
	}

	log.Info(fmt.Sprintf("SecretScope not found: %s", s))

	return SecretScope("unknown")
}

// SecretResolver resolver for Secret
type SecretResolver struct {
	Secret
	SecretValue SecretValue
	DB          *gorm.DB
}

// ID
func (r *SecretResolver) ID() graphql.ID {
	return graphql.ID(r.Secret.Model.ID.String())
}

// Key
func (r *SecretResolver) Key() string {
	return r.Secret.Key
}

// Value
func (r *SecretResolver) Value() string {
	if r.IsSecret() {
		return ""
	}

	if r.SecretValue != (SecretValue{}) {
		return r.SecretValue.Value
	} else {
		return r.Secret.Value.Value
	}
}

// Scope
func (r *SecretResolver) Scope() string {
	return string(r.Secret.Scope)
}

// Project
func (r *SecretResolver) Project() *ProjectResolver {
	var project Project

	r.DB.Model(r.Secret).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// User
func (r *SecretResolver) User() *UserResolver {
	var user User

	r.DB.Model(r.SecretValue).Related(&user)

	return &UserResolver{DB: r.DB, User: user}
}

// Type
func (r *SecretResolver) Type() string {
	return string(r.Secret.Type)
}

// Versions
func (r *SecretResolver) Versions() ([]*SecretResolver, error) {
	var secretValues []SecretValue
	var secretResolvers []*SecretResolver

	r.DB.Where("secret_id = ?", r.Secret.Model.ID).Order("created_at desc").Find(&secretValues)

	for _, secretValue := range secretValues {
		secretResolvers = append(secretResolvers, &SecretResolver{DB: r.DB, Secret: r.Secret, SecretValue: secretValue})
	}

	return secretResolvers, nil
}

// Environment
func (r *SecretResolver) Environment() *EnvironmentResolver {
	var env Environment

	r.DB.Model(r.Secret).Related(&env)

	return &EnvironmentResolver{DB: r.DB, Environment: env}
}

// Created
func (r *SecretResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Secret.Model.CreatedAt}
}

// IsSecret
func (r *SecretResolver) IsSecret() bool {
	return r.Secret.IsSecret
}

func (r *SecretResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Secret)
}

func (r *SecretResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Secret)
}
