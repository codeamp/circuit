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
	return ``, nil
}
