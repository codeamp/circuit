package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type Service struct {
	*model.Service
}

type ServiceConfig struct {
	ProjectConfig
	service *model.Service
}

func CreateServiceConfig(config *string, db *gorm.DB, service *model.Service, project *model.Project, env *model.Environment) *ServiceConfig {
	return &ServiceConfig{
		service: service,
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

func (s *ServiceConfig) Export() (*Service, error) {
	if s.service == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "service")
	}

	return &Service{
		Service: s.service,
	}, nil
}
