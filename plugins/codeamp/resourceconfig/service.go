package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectService struct{}

type ServiceConfig struct{ ProjectConfig }

func CreateServiceConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *ServiceConfig {
	return &ServiceConfig{
		ProjectConfig: ProjectConfig{
			db:          db,
			project:     project,
			environment: env,
			BaseResourceConfig: BaseResourceConfig{
				config: config,
			},
		},
	}
}

func (s *ServiceConfig) ExportYAML() (string, error) {
	aggregateConfigString := ``
	// marshal serivce yaml first

	childConfigs, err := s.GetChildResourceConfigs()
	if err != nil {
		return ``, err
	}

	var configString string
	for _, config := range childConfigs {
		configString, err = config.ExportYAML()
		if err != nil {
			return ``, err
		}

		aggregateConfigString += configString
	}

	return aggregateConfigString, nil
}
