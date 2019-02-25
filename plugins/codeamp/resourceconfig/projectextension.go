package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ProjectExtensionConfig struct {
	projectExtension *model.ProjectExtension
	ProjectConfig
}

type Extension struct {
	Key string `yaml:"name"`
}

type ProjectExtension struct {
	CustomConfig []byte `yaml:"customConfig"`
	Config       []byte `yaml:"config"`
	Key          string `yaml:"key"`
}

func CreateProjectExtensionConfig(config *string, db *gorm.DB, projectExtension *model.ProjectExtension, project *model.Project, env *model.Environment) *ProjectExtensionConfig {
	return &ProjectExtensionConfig{
		projectExtension: projectExtension,
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

func (p *ProjectExtensionConfig) Export() (*ProjectExtension, error) {
	extension := model.Extension{}

	if p.projectExtension == nil || p.db == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "projectExtension, db")
	}

	if err := p.db.Where("id = ?", p.projectExtension.ExtensionID).First(&extension).Error; err != nil {
		return nil, err
	}

	return &ProjectExtension{
		CustomConfig: p.projectExtension.CustomConfig.RawMessage,
		Config:       p.projectExtension.Config.RawMessage,
		Key:          extension.Key,
	}, nil
}
