package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectSettingsConfig struct {
	ProjectConfig
	projectSettings *model.ProjectSettings
}

type ProjectSettings struct {
	GitBranch        string `yaml:"gitBranch"`
	ContinuousDeploy bool   `yaml:"continuousDeploy"`
}

func CreateProjectSettingsConfig(config *string, db *gorm.DB, projectSettings *model.ProjectSettings, project *model.Project, env *model.Environment) *ProjectSettingsConfig {
	return &ProjectSettingsConfig{
		projectSettings: projectSettings,
		ProjectConfig: ProjectConfig{
			db:                 db,
			project:            project,
			environment:        env,
			BaseResourceConfig: BaseResourceConfig{config: config},
		},
	}
}

func (p *ProjectSettingsConfig) Export() (*ProjectSettings, error) {
	if p.projectSettings != nil {
		return &ProjectSettings{
			GitBranch:        p.projectSettings.GitBranch,
			ContinuousDeploy: p.projectSettings.ContinuousDeploy,
		}, nil
	} else {
		return nil, fmt.Errorf(NilDependencyForExportErr, "projectSettings")
	}
}
