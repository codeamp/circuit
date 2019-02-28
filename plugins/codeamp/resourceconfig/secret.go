package resourceconfig

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/helpers"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type SecretConfig struct {
	BaseResourceConfig
	db          *gorm.DB
	secret      *model.Secret
	project     *model.Project
	environment *model.Environment
	ctx         context.Context // used for getting the user id when importing
}

type Secret struct {
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
	Type     string `yaml:"type"`
	IsSecret bool   `yaml:"isSecret"`
}

func CreateProjectSecretConfig(ctx context.Context, db *gorm.DB, secret *model.Secret, project *model.Project, env *model.Environment) *SecretConfig {
	return &SecretConfig{
		db:          db,
		secret:      secret,
		project:     project,
		ctx:         ctx,
		environment: env,
	}
}

func (p *SecretConfig) Import(secret *Secret) error {
	if p.db == nil || p.project == nil || p.environment == nil {
		return fmt.Errorf(NilDependencyForExportErr, "db, project, environment")
	}

	projectID := p.project.Model.ID.String()

	secretInput := model.SecretInput{
		Key:           secret.Key,
		Value:         secret.Value,
		Type:          secret.Type,
		Scope:         "project",
		ProjectID:     &projectID,
		EnvironmentID: p.environment.Model.ID.String(),
		IsSecret:      secret.IsSecret,
	}

	_, err := helpers.CreateSecretInDB(p.ctx, p.db, &secretInput)
	if err != nil {
		return err
	}

	return nil
}

func (s *SecretConfig) Export() (*Secret, error) {
	if s.secret == nil || s.db == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "secret, db")
	}

	secretValue := model.SecretValue{}
	if err := s.db.Where("secret_id = ?", s.secret.Model.ID).Order("created_at desc").Find(&secretValue).Error; err != nil {
		return nil, err
	}

	return &Secret{
		Key:      s.secret.Key,
		Value:    secretValue.Value,
		Type:     string(s.secret.Type),
		IsSecret: s.secret.IsSecret,
	}, nil
}
