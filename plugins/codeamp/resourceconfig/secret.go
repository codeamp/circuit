package resourceconfig

import (
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
	return &Secret{}, nil
}
