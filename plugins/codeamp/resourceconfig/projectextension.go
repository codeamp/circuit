package resourceconfig

import (
	"encoding/json"
	"fmt"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

type ProjectExtensionConfig struct {
	projectExtension *model.ProjectExtension
	ProjectConfig
}

type ProjectExtension struct {
	CustomConfig string `yaml:"customConfig"`
	Config       string `yaml:"config"`
	Key          string `yaml:"key"`
}

func CreateProjectExtensionConfig(db *gorm.DB, projectExtension *model.ProjectExtension, project *model.Project, env *model.Environment) *ProjectExtensionConfig {
	return &ProjectExtensionConfig{
		projectExtension: projectExtension,
		ProjectConfig: ProjectConfig{
			db:          db,
			project:     project,
			environment: env,
		},
	}
}

func (p *ProjectExtensionConfig) Import(projectExtension *ProjectExtension) error {
	if projectExtension == nil || p.db == nil || p.project == nil || p.environment == nil {
		return fmt.Errorf(NilDependencyForExportErr, "projectExtension, db, project, environment")
	}

	// convert to project extension input

	return nil
}

func (p *ProjectExtensionConfig) Export() (*ProjectExtension, error) {
	extension := model.Extension{}

	if p.projectExtension == nil || p.db == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "projectExtension, db")
	}

	if err := p.db.Where("id = ?", p.projectExtension.ExtensionID).First(&extension).Error; err != nil {
		return nil, err
	}

	configArtifacts := []transistor.Artifact{}

	unmarshaledProjectExtensionConfig := []transistor.Artifact{}
	err := json.Unmarshal(p.projectExtension.Config.RawMessage, &unmarshaledProjectExtensionConfig)
	if err != nil {
		return nil, err
	}

	for _, artifact := range unmarshaledProjectExtensionConfig {
		secretValue := model.SecretValue{}
		// config artifact value is a reference to the actual secret object
		secretID := artifact.Value
		if p.db.Where("secret_id = ?", secretID).Order("created_at desc").First(&secretValue).RecordNotFound() {
			log.InfoWithFields("secret value not found", log.Fields{
				"secret_id": secretID,
			})
		}

		configArtifact := transistor.Artifact{
			Key:   artifact.Key,
			Value: secretValue.Value,
		}
		configArtifacts = append(configArtifacts, configArtifact)
	}

	artifactsBytes, err := yaml.Marshal(configArtifacts)
	if err != nil {
		return nil, err
	}

	return &ProjectExtension{
		CustomConfig: string(p.projectExtension.CustomConfig.RawMessage),
		Config:       string(artifactsBytes),
		Key:          extension.Key,
	}, nil
}
