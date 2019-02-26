package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/helpers"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

type ProjectExtensionConfig struct {
	projectExtension *model.ProjectExtension
	ProjectConfig
}

type Extension struct {
	Key string `yaml:"name"`
}

type ProjectExtension struct {
	CustomConfig string `yaml:"customConfig"`
	Config       string `yaml:"config"`
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

	artifacts, err := helpers.ExtractArtifacts(*p.projectExtension, extension, p.db)
	if err != nil {
		return nil, err
	}

	artifactsBytes, err := yaml.Marshal(artifacts)
	if err != nil {
		return nil, err
	}

	return &ProjectExtension{
		CustomConfig: string(p.projectExtension.CustomConfig.RawMessage),
		Config:       string(artifactsBytes),
		Key:          extension.Key,
	}, nil
}
