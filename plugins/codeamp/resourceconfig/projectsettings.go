package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectSettingsConfig struct {
	ProjectConfig
}

type ProjectSettings struct {
	GitBranch        string `yaml:"gitBranch"`
	ContinuousDeploy bool   `yaml:"continuousDeploy"`
}

func CreateProjectSettingsConfig(db *gorm.DB, project *model.Project, env *model.Environment) *ProjectSettingsConfig {
	return &ProjectSettingsConfig{
		ProjectConfig: ProjectConfig{
			db:          db,
			project:     project,
			environment: env,
		},
	}
}

func (p *ProjectSettingsConfig) Import(settings *ProjectSettings) error {
	if p.db == nil || p.project == nil || p.environment == nil {
		return fmt.Errorf(NilDependencyForExportErr, "db, project, environment")
	}

	// get project settings for this
	dbProjectSettings := model.ProjectSettings{}
	if err := p.db.Where("project_id = ? and environment_id = ?", p.project.Model.ID, p.environment.Model.ID).Find(&dbProjectSettings).Error; err != nil {
		return err
	}

	// update project settings
	dbProjectSettings.GitBranch = settings.GitBranch
	dbProjectSettings.ContinuousDeploy = settings.ContinuousDeploy

	if err := p.db.Save(&dbProjectSettings).Error; err != nil {
		return err
	}

	return nil
}

func (p *ProjectSettingsConfig) Export() (*ProjectSettings, error) {
	if p.db == nil || p.project == nil || p.environment == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "db, project, environment")
	}

	dbProjectSettings := model.ProjectSettings{}
	if err := p.db.Where("project_id = ? and environment_id  = ?", p.project.Model.ID, p.environment.Model.ID).Find(&dbProjectSettings).Error; err != nil {
		return nil, err
	}

	return &ProjectSettings{
		GitBranch:        dbProjectSettings.GitBranch,
		ContinuousDeploy: dbProjectSettings.ContinuousDeploy,
	}, nil
}
