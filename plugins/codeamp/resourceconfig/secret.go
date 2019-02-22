package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

type SecretConfig struct {
	BaseResourceConfig
	db          *gorm.DB
	project     *model.Project
	environment *model.Environment
}

type Secret struct {
	Key   string `yaml:"key"`
	Value string `yaml:"value"`
}

func CreateSecretConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *SecretConfig {
	return &SecretConfig{
		db:                 db,
		project:            project,
		environment:        env,
		BaseResourceConfig: BaseResourceConfig{config: config},
	}
}

func (c *SecretConfig) ExportYAML() (string, error) {
	secret := Secret{}

	secretYamlString, err := yaml.Marshal(secret)
	if err != nil {
		return ``, nil
	}

	return string(secretYamlString), nil
}
