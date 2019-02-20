package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectExtensionConfig struct{ ProjectConfig }

type ProjectExtension struct {
}

func CreateProjectExtensionConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *ProjectExtensionConfig {
	return &ProjectExtensionConfig{
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
