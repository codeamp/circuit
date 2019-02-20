package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type SecretConfig struct {
	BaseResourceConfig
	db          *gorm.DB
	project     *model.Project
	environment *model.Environment
}

type Secret struct{}

func CreateSecretConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *SecretConfig {
	return &SecretConfig{
		db:                 db,
		project:            project,
		environment:        env,
		BaseResourceConfig: BaseResourceConfig{config: config},
	}
}
