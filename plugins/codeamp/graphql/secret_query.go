package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	yaml "gopkg.in/yaml.v2"
)

// Secret Resolver Query
type SecretResolverQuery struct {
	DB *gorm.DB
}

func (r *SecretResolverQuery) Secrets(ctx context.Context, args *struct {
	Params *model.PaginatorInput
}) (*SecretListResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	db := r.DB.Where("scope != ?", "project").Order("environment_id desc, key asc, scope asc")
	return &SecretListResolver{
		DBSecretListResolver: &db_resolver.SecretListResolver{
			DB:             db,
			PaginatorInput: args.Params,
		},
	}, nil
}

func (r *SecretResolverQuery) Secret(ctx context.Context, args *struct {
	ID *string
}) (*SecretResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	isAdmin := false
	if userID, err := auth.CheckAuth(ctx, []string{"admin"}); err != nil {
		log.Error(err.Error())
	} else {
		if userID != "" {
			isAdmin = true
		}
	}
	resolver := SecretResolver{DBSecretResolver: &db_resolver.SecretResolver{DB: r.DB, IsAdmin: isAdmin}}

	r.DB.Where("id = ?", args.ID).First(&resolver.DBSecretResolver.Secret)
	return &resolver, nil
}

// ExportSecrets returns a list of all secrets for a given project and environment in a YAML string format
func (r *SecretResolverQuery) ExportSecrets(ctx context.Context, args *struct{ Params *model.ExportSecretsInput }) (string, error) {
	project := model.Project{}
	env := model.Environment{}
	secrets := []model.Secret{}

	if err := r.DB.Where("id = ?", args.Params.ProjectID).First(&project).Error; err != nil {
		return "", err
	}

	if err := r.DB.Where("id = ?", args.Params.EnvironmentID).First(&env).Error; err != nil {
		return "", err
	}

	if err := r.DB.Where("project_id = ? and environment_id = ?", project.Model.ID.String(), env.Model.ID.String()).Find(&secrets).Error; err != nil {
		return "", err
	}

	// convert to YAMLSecret and then marshal
	yamlSecrets := []model.YAMLSecret{}
	for _, secret := range secrets {
		value := secret.Value.Value
		if secret.IsSecret {
			value = ""
		}

		yamlSecrets = append(yamlSecrets, model.YAMLSecret{
			Key:      secret.Key,
			Value:    value,
			IsSecret: secret.IsSecret,
			Type:     string(secret.Type),
		})
	}

	out, err := yaml.Marshal(yamlSecrets)
	if err != nil {
		return "", err
	}

	return string(out), nil
}
