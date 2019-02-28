package resourceconfig

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type SecretConfig struct {
	BaseResourceConfig
	db          *gorm.DB
	secret      *model.Secret
	project     *model.Project
	environment *model.Environment
}

type Secret struct {
	Key      string `yaml:"key"`
	Value    string `yaml:"value"`
	Type     string `yaml:"type"`
	IsSecret bool   `yaml:"isSecret"`
}

func CreateProjectSecretConfig(db *gorm.DB, secret *model.Secret, project *model.Project, env *model.Environment) *SecretConfig {
	return &SecretConfig{
		db:          db,
		secret:      secret,
		project:     project,
		environment: env,
	}
}

func (p *SecretConfig) Import(secret *Secret) error {
	if p.db == nil || p.project == nil || p.environment == nil {
		return fmt.Errorf(NilDependencyForExportErr, "db, project, environment")
	}

	// check if secret already exists
	if err := p.db.Where("project_id = ? and environment_id = ? and key = ?", p.project.Model.ID, p.environment.Model.ID, secret.Key).Find(&model.Secret{}).Error; err == nil {
		return fmt.Errorf(ObjectAlreadyExistsErr, "Secret")
	}

	newDBSecret := model.Secret{
		Key: secret.Key,
	}
	if err := p.db.Create(&newDBSecret).Error; err != nil {
		return err
	}

	newDBSecretValue := model.SecretValue{
		Value:    secret.Value,
		SecretID: newDBSecret.Model.ID,
	}
	if err := p.db.Create(&newDBSecretValue).Error; err != nil {
		return err
	}

	return nil
}

func (s *SecretConfig) Export() (*Secret, error) {
	if s.secret == nil || s.db == nil {
		return nil, fmt.Errorf(NilDependencyForExportErr, "secret, db")
	}

	secretValue := model.SecretValue{}
	if err := s.db.Where("secret_id = ?", s.secret.Model.ID).Order("created_at desc").Find(&secretValue).Error; err != nil {
		return nil, err
	}

	return &Secret{
		Key:      s.secret.Key,
		Value:    secretValue.Value,
		Type:     string(s.secret.Type),
		IsSecret: s.secret.IsSecret,
	}, nil
}
