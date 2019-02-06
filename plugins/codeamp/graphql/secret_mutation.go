package graphql_resolver

import (
	"context"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	yaml "gopkg.in/yaml.v2"
)

// Secret Resolver Mutation
type SecretResolverMutation struct {
	DB *gorm.DB
}

func (r *SecretResolverMutation) CreateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {

	projectID := uuid.UUID{}
	var environmentID uuid.UUID
	var secretScope model.SecretScope

	if args.Secret.ProjectID != nil {
		// Check if project can create secret
		if r.DB.Where("environment_id = ? and project_id = ?", args.Secret.EnvironmentID, args.Secret.ProjectID).Find(&model.ProjectEnvironment{}).RecordNotFound() {
			return nil, errors.New("Project not allowed to create secret in given environment")
		}

		projectID = uuid.FromStringOrNil(*args.Secret.ProjectID)
	}

	secretScope = GetSecretScope(args.Secret.Scope)
	if secretScope == model.SecretScope("unknown") {
		return nil, fmt.Errorf("Invalid env var scope.")
	}

	environmentID, err := uuid.FromString(args.Secret.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse environmentID. Invalid format.")
	}

	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return &SecretResolver{}, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return &SecretResolver{}, err
	}

	var existingEnvVar model.Secret

	if r.DB.Where("key = ? and project_id = ? and deleted_at is null and environment_id = ? and type = ?", args.Secret.Key, projectID, environmentID, args.Secret.Type).Find(&existingEnvVar).RecordNotFound() {
		secret := model.Secret{
			Key:           args.Secret.Key,
			ProjectID:     projectID,
			Type:          plugins.GetType(args.Secret.Type),
			Scope:         secretScope,
			EnvironmentID: environmentID,
			IsSecret:      args.Secret.IsSecret,
		}
		r.DB.Create(&secret)

		secretValue := model.SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretCreated(&secret)

		return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}}, nil
	} else {
		return nil, fmt.Errorf("CreateSecret: key already exists")
	}

}

func (r *SecretResolverMutation) UpdateSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	var secret model.Secret

	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return &SecretResolver{}, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return &SecretResolver{}, err
	}

	if r.DB.Where("id = ?", args.Secret.ID).Find(&secret).RecordNotFound() {
		return nil, fmt.Errorf("UpdateSecret: env var doesn't exist.")
	} else {
		secretValue := model.SecretValue{
			SecretID: secret.Model.ID,
			Value:    args.Secret.Value,
			UserID:   userID,
		}
		r.DB.Create(&secretValue)

		//r.SecretUpdated(&secret)

		return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret, SecretValue: secretValue}}, nil
	}
}

func (r *SecretResolverMutation) DeleteSecret(ctx context.Context, args *struct{ Secret *model.SecretInput }) (*SecretResolver, error) {
	var secret model.Secret

	if r.DB.Where("id = ?", args.Secret.ID).Find(&secret).RecordNotFound() {
		return nil, fmt.Errorf("DeleteSecret: key doesn't exist.")
	} else {
		// check if any configs are using the secret
		extensions := []model.Extension{}
		where := fmt.Sprintf(`config @> '{"config": [{"value": "%s"}]}'`, secret.Model.ID.String())
		r.DB.Where(where).Find(&extensions)
		if len(extensions) == 0 {
			versions := []model.SecretValue{}

			r.DB.Delete(&secret)
			r.DB.Where("secret_id = ?", secret.Model.ID).Delete(&versions)

			//r.SecretDeleted(&secret)

			return &SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, Secret: secret}}, nil
		} else {
			return nil, fmt.Errorf("Remove Config values from Extensions where Secret is used before deleting.")
		}
	}
}

func (r *SecretResolverMutation) ImportSecrets(ctx context.Context, args *struct{ Secrets *model.ImportSecretsInput }) ([]*SecretResolver, error) {
	importedSecrets := []model.YAMLSecret{}
	createdSecrets := []*SecretResolver{}

	err := yaml.Unmarshal([]byte(args.Secrets.SecretsYAMLString), &importedSecrets)
	if err != nil {
		return nil, err
	}

	project := model.Project{}
	if err := r.DB.Where("id = ?", args.Secrets.ProjectID).First(&project).Error; err != nil {
		return nil, err
	}

	env := model.Environment{}
	if err := r.DB.Where("id = ?", args.Secrets.EnvironmentID).First(&env).Error; err != nil {
		return nil, err
	}

	user := model.User{}
	userIDString, err := auth.CheckAuth(ctx, []string{})
	if err != nil {
		return nil, err
	}

	userID, err := uuid.FromString(userIDString)
	if err != nil {
		return nil, err
	}

	if err := r.DB.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, err
	}

	tx := r.DB.Begin()
	// create secrets only if they're new
	for _, importedSecret := range importedSecrets {
		// check if key already exists in this project, environment
		existing := model.Secret{}
		importedSecretType := plugins.GetType(importedSecret.Type)
		newSecret := model.Secret{
			Key:           importedSecret.Key,
			Type:          importedSecretType,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			IsSecret:      importedSecret.IsSecret,
		}
		newSecretValue := model.SecretValue{
			Value:  importedSecret.Value,
			UserID: user.Model.ID,
		}

		if importedSecretType == plugins.Type("unknown") {
			return nil, fmt.Errorf("Invalid type for secret key %s", importedSecret.Key)
		}

		if err := tx.Where("project_id = ? and environment_id = ? and key = ?", project.Model.ID, env.Model.ID, importedSecret.Key).First(&existing).Error; err == nil {
			log.InfoWithFields("Secret already exists", log.Fields{
				"key":        importedSecret.Key,
				"project_id": project.Model.ID,
				"secret_id":  env.Model.ID,
			})
		} else {
			if err := tx.Create(&newSecret).Error; err != nil {
				return nil, err
			}

			newSecretValue.SecretID = newSecret.Model.ID
			if err := tx.Create(&newSecretValue).Error; err != nil {
				return nil, err
			}

			createdSecrets = append(createdSecrets, &SecretResolver{
				DBSecretResolver: &db_resolver.SecretResolver{
					Secret:      newSecret,
					SecretValue: newSecretValue,
					DB:          r.DB,
				},
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.Error(err.Error())
		return nil, err
	}

	return createdSecrets, nil
}
