package resourceconfig

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
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
	ProjectSettings   ProjectSettings    `yaml:"settings`
	ProjectExtensions []ProjectExtension `yaml:"extensions"`
	Secrets           []Secret           `yaml"secrets"`
	Services          []ProjectService   `yaml:"servies"`
}

func CreateProjectConfig(config string, db *gorm.DB, project *model.Project, env *model.Environment) *ProjectConfig {
	return &ProjectConfig{
		db:          db,
		project:     project,
		environment: env,
	}
}

func (p *ProjectConfig) ExportYAML() (string, error) {
	aggregateConfigString := ``
	childConfigs, err := p.GetChildResourceConfigs()
	if err != nil {
		return ``, err
	}

	var configString string
	for _, config := range childConfigs {
		configString, err = config.ExportYAML()
		if err != nil {
			return ``, err
		}

		aggregateConfigString += configString
	}

	// test if can be unmarshaled
	unmarshaledProject := Project{}
	err = yaml.Unmarshal([]byte(aggregateConfigString), &unmarshaledProject)
	if err != nil {
		return ``, err
	}

	return aggregateConfigString, nil
}

func (p *ProjectConfig) GetChildResourceConfigs() ([]ResourceConfig, error) {
	childResourceConfigs := []ResourceConfig{}
	// unmarshal config and get resource configs for each child object
	project := model.Project{}
	err := yaml.Unmarshal([]byte(p.GetConfig()), &project)
	if err != nil {
		return nil, err
	}

	for _, service := range services {
		childResourceConfigs = append(childResourceConfigs, CreateServiceConfig(``, p.db, p.project, p.environment))
	}

	for _, extension := range project.ProjectExtensions {
		childResourceConfigs = append(childResourceConfigs, CreateProjectExtensionConfig(``, p.db, p.project, p.environment))
	}

	for _, secret := range project.Secrets {
		childResourceConfigs = append(childResourceConfigs, CreateSecretConfig(``, p.db, p.project, p.environment))
	}

	childResourceConfigs = append(childResourceConfigs, CreateProjectSettingsConfig(``, p.db, p.project, p.environment))

	return childResourceConfigs, nil
}
