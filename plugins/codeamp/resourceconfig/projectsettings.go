package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectSettingsConfig struct {
	ProjectConfig
}

type ProjectSettings struct {
}

func CreateProjectSettingsConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *ProjectSettingsConfig {
	return &ProjectSettingsConfig{
		ProjectConfig: ProjectConfig{
			db:                 db,
			project:            project,
			environment:        env,
			BaseResourceConfig: BaseResourceConfig{config: config},
		},
	}
}
