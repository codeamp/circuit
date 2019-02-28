package helpers

import (
	"context"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

func CreateSecretInDB(ctx context.Context, tx *gorm.DB, input *model.SecretInput) (*model.Secret, error) {
	projectID := uuid.UUID{}
	var environmentID uuid.UUID
	var secretScope model.SecretScope

	if input.ProjectID != nil {
		// Check if project can create secret
		if tx.Where("environment_id = ? and project_id = ?", input.EnvironmentID, input.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
			return nil, errors.New("Project not allowed to create secret in given environment")
		}

		projectID = uuid.FromStringOrNil(*input.ProjectID)
	}

	secretScope = GetSecretScope(input.Scope)
	if secretScope == model.SecretScope("unknown") {
		return nil, fmt.Errorf("Invalid env var scope.")
	}

	environmentID, err := uuid.FromString(input.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse environmentID. Invalid format.")
	}

	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return nil, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return nil, err
	}

	var existingEnvVar model.Secret

	if tx.Where("key = ? and project_id = ? and deleted_at is null and environment_id = ? and type = ?", input.Key, projectID, environmentID, input.Type).Find(&existingEnvVar).RecordNotFound() {
		secret := model.Secret{
			Key:           input.Key,
			ProjectID:     projectID,
			Type:          plugins.GetType(input.Type),
			Scope:         secretScope,
			EnvironmentID: environmentID,
			IsSecret:      input.IsSecret,
		}
		tx.Create(&secret)

		secretValue := model.SecretValue{
			SecretID: secret.Model.ID,
			Value:    input.Value,
			UserID:   userID,
		}
		tx.Create(&secretValue)

		return &secret, nil
	}

	return nil, fmt.Errorf("secret found with key %s", input.Key)
}

func GetSecretScope(s string) model.SecretScope {
	secretScopes := []string{
		"project",
		"extension",
		"global",
	}

	for _, secretScope := range secretScopes {
		if s == secretScope {
			return model.SecretScope(secretScope)
		}
	}

	log.Warn(fmt.Sprintf("SecretScope not found: %s", s))
	return model.SecretScope("unknown")
}
