package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type SecretConfig struct {
	BaseResourceConfig
	db          *gorm.DB
	secret      *model.Secret
	project     *model.Project
	environment *model.Environment
}

type Secret struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func CreateSecretConfig(config *string, db *gorm.DB, secret *model.Secret, project *model.Project, env *model.Environment) *SecretConfig {
	return &SecretConfig{
		db:                 db,
		secret:             secret,
		project:            project,
		environment:        env,
		BaseResourceConfig: BaseResourceConfig{config: config},
	}
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
		Key:   s.secret.Key,
		Value: secretValue.Value,
	}, nil
}
