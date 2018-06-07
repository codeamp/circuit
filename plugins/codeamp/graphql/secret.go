package graphql_resolver

import (
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

func GetSecretScope(s string) model.SecretScope {
	secretScopes := []string{
		"project",
		"extension",
		"global",
	}

	for _, secretScope := range secretScopes {
		if s == secretScope {
			return model.SecretScope(secretScope)
		}
	}

	log.Info(fmt.Sprintf("SecretScope not found: %s", s))

	return model.SecretScope("unknown")
}

// SecretResolver resolver for Secret
type SecretResolver struct {
	model.Secret
	SecretValue model.SecretValue
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

	if r.SecretValue != (model.SecretValue{}) {
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
	var project model.Project

	r.DB.Model(r.Secret).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// User
func (r *SecretResolver) User() *UserResolver {
	var user model.User

	r.DB.Model(r.SecretValue).Related(&user)

	// return &UserResolver{DB: r.DB, User: user}
	log.Panic("PANIC")
	return nil
}

// Type
func (r *SecretResolver) Type() string {
	return string(r.Secret.Type)
}

// Versions
func (r *SecretResolver) Versions() ([]*SecretResolver, error) {
	var secretValues []model.SecretValue
	var secretResolvers []*SecretResolver

	r.DB.Where("secret_id = ?", r.Secret.Model.ID).Order("created_at desc").Find(&secretValues)

	for _, secretValue := range secretValues {
		secretResolvers = append(secretResolvers, &SecretResolver{DB: r.DB, Secret: r.Secret, SecretValue: secretValue})
	}

	return secretResolvers, nil
}

// Environment
func (r *SecretResolver) Environment() *EnvironmentResolver {
	var env model.Environment

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
