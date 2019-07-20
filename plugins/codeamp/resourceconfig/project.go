package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// example struct that implements ResourceConfig
type ProjectConfig struct {
	BaseResourceConfig // inherits methods from BaseConfig
	db                 *gorm.DB
	project            *model.Project
	environment        *model.Environment
}

// For exporting purposes
type Project struct {
	ProjectSettings   ProjectSettings    `yaml:"settings"`
	ProjectExtensions []ProjectExtension `yaml:"extensions"`
	Secrets           []Secret           `yaml:"secrets"`
	Services          []Service          `yaml:"services"`
}

func CreateProjectConfig(config *string, db *gorm.DB, project *model.Project, env *model.Environment) *ProjectConfig {
	return &ProjectConfig{
		db:          db,
		project:     project,
		environment: env,
		BaseResourceConfig: BaseResourceConfig{
			config: config,
		},
	}
}

func (p *ProjectConfig) Export() (*Project, error) {
	var project Project

	if p.project == nil || p.environment == nil || p.db == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "project, environment, db")
	}

	childObjectQuery := p.db.Where("project_id = ? and environment_id = ?", p.project.Model.ID, p.environment.Model.ID)

	// Collect services inside project
	var services []model.Service
	if err := childObjectQuery.Find(&services).Error; err != nil {
		return nil, err
	}

	for _, service := range services {
		exportedSvc, err := CreateServiceConfig(nil, p.db, &service, nil, nil).Export()
		if err != nil {
			return nil, err
		}

		project.Services = append(project.Services, *exportedSvc)
	}

	// Collect project extensions inside project
	var pExtensions []model.ProjectExtension
	if err := childObjectQuery.Find(&pExtensions).Error; err != nil {
		return nil, err
	}

	for _, pExtension := range pExtensions {
		exportedProjectExtension, err := CreateProjectExtensionConfig(nil, p.db, &pExtension, nil, nil).Export()
		if err != nil {
			return nil, err
		}

		project.ProjectExtensions = append(project.ProjectExtensions, *exportedProjectExtension)
	}

	var secrets []model.Secret
	if err := childObjectQuery.Find(&secrets).Error; err != nil {
		return nil, err
	}

	// Collect services inside project
	for _, secret := range secrets {
		exportedSecret, err := CreateSecretConfig(nil, p.db, &secret, nil, nil).Export()
		if err != nil {
			return nil, err
		}

		project.Secrets = append(project.Secrets, *exportedSecret)
	}

	projectSettings := model.ProjectSettings{}
	if err := childObjectQuery.Find(&projectSettings).Error; err != nil {
		return nil, err
	}

	exportedProjectSettings, err := CreateProjectSettingsConfig(nil, p.db, &projectSettings, nil, nil).Export()
	if err != nil {
		return nil, err
	}

	project.ProjectSettings = *exportedProjectSettings

	return &project, nil
}
