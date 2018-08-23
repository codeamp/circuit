package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type SecretResolver struct {
	model.Secret
	model.SecretValue
	DB *gorm.DB
}

// IsSecret
func (r *SecretResolver) IsSecret() bool {
	return r.Secret.IsSecret
}

// Value
func (r *SecretResolver) Value() string {
	if r.IsSecret() {
		return "******"
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
func (r *SecretResolver) User() (*UserResolver, error) {
	// Find the least most recent secret
	if r.SecretValue == (model.SecretValue{}) {
		if err := r.DB.Where("secret_id = ?", r.Secret.Model.ID).Order("created_at asc").Find(&r.SecretValue).Error; err != nil {
			return nil, err
		}
	}

	var user model.User
	if err := r.DB.Where("id = ?", r.SecretValue.UserID).Find(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{DB: r.DB, User: user}, nil
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
