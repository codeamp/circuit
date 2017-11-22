package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/davecgh/go-spew/spew"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type EnvironmentVariableInput struct {
	ID            *string
	Key           string
	Value         string
	Type          string
	Scope         string
	ProjectId     *string
	EnvironmentId string
}

type EnvironmentVariableResolver struct {
	db                  *gorm.DB
	EnvironmentVariable models.EnvironmentVariable
}

func (r *Resolver) EnvironmentVariable(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentVariableResolver, error) {
	envVar := models.EnvironmentVariable{}
	if err := r.db.Where("id = ?", args.ID).First(&envVar).Error; err != nil {
		return nil, err
	}
	spew.Dump(args)
	spew.Dump(envVar)

	return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
}

func (r *Resolver) CreateEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {

	projectId := uuid.UUID{}
	var environmentId uuid.UUID
	var envVarScope plugins.EnvVarScope

	spew.Dump(args.EnvironmentVariable.ProjectId)

	if args.EnvironmentVariable.ProjectId != nil {
		projectId = uuid.FromStringOrNil(*args.EnvironmentVariable.ProjectId)
		envVarScope = plugins.EnvVarScope(args.EnvironmentVariable.Scope)
	} else {
		envVarScope = plugins.GlobalScope
	}

	environmentId, err := uuid.FromString(args.EnvironmentVariable.EnvironmentId)
	if err != nil {
		return nil, fmt.Errorf("Couldn't parse environmentId. Invalid format.")
	}

	userIdString, err := utils.CheckAuth(ctx, []string{})
	if err != nil {
		return &EnvironmentVariableResolver{}, err
	}

	userId, err := uuid.FromString(userIdString)
	if err != nil {
		return &EnvironmentVariableResolver{}, err
	}

	var existingEnvVar models.EnvironmentVariable

	if r.db.Where("key = ? and project_id = ? and deleted_at is null and environment_id = ?", args.EnvironmentVariable.Key, projectId, environmentId).Find(&existingEnvVar).RecordNotFound() {
		envVar := models.EnvironmentVariable{
			Key:           args.EnvironmentVariable.Key,
			Value:         args.EnvironmentVariable.Value,
			ProjectId:     projectId,
			Version:       int32(0),
			Type:          plugins.Type(args.EnvironmentVariable.Type),
			Scope:         envVarScope,
			UserId:        userId,
			EnvironmentId: environmentId,
		}

		r.db.Create(&envVar)

		r.actions.EnvironmentVariableCreated(&envVar)

		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironmentVariable: key already exists")
	}
}

func (r *Resolver) UpdateEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {

	var existingEnvVar models.EnvironmentVariable
	environmentId := uuid.FromStringOrNil(args.EnvironmentVariable.EnvironmentId)

	if r.db.Where("id = ?", args.EnvironmentVariable.ID).Find(&existingEnvVar).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnvironmentVariable: env var doesn't exist.")
	} else {
		envVar := models.EnvironmentVariable{
			Key:           args.EnvironmentVariable.Key,
			Value:         args.EnvironmentVariable.Value,
			ProjectId:     existingEnvVar.ProjectId,
			Version:       existingEnvVar.Version + int32(1),
			Type:          existingEnvVar.Type,
			Scope:         plugins.EnvVarScope(args.EnvironmentVariable.Scope),
			UserId:        existingEnvVar.UserId,
			EnvironmentId: environmentId,
		}
		r.db.Delete(&existingEnvVar)
		r.db.Create(&envVar)

		// find all extension specs using the env var if project id is nil
		if envVar.Scope != plugins.ProjectScope {
			var extensionSpecEnvVars []models.ExtensionSpecEnvironmentVariable
			if r.db.Where("environment_variable_id = ?", args.EnvironmentVariable.ID).Find(&extensionSpecEnvVars).RecordNotFound() {
				log.InfoWithFields("Nothing to update", log.Fields{
					"envVar": envVar,
				})
			}
			for _, extensionSpecEnvVar := range extensionSpecEnvVars {
				extensionSpecEnvVar.EnvironmentVariableId = envVar.Model.ID
				r.db.Save(&extensionSpecEnvVar)
			}
		}

		r.actions.EnvironmentVariableUpdated(&envVar)

		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
	}
}

func (r *Resolver) DeleteEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {

	var existingEnvVar models.EnvironmentVariable
	if r.db.Where("id = ?", args.EnvironmentVariable.ID).Find(&existingEnvVar).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnvironmentVariable: key doesn't exist.")
	} else {
		var rows []models.EnvironmentVariable

		r.db.Where("project_id = ? and key = ? and environment_id = ?", existingEnvVar.ProjectId, existingEnvVar.Key, existingEnvVar.EnvironmentId).Find(&rows)
		spew.Dump("FOUND EM", rows)
		for _, envVar := range rows {
			r.db.Unscoped().Delete(&envVar)
		}

		// find all extension specs using the env var if project id is nil
		if existingEnvVar.Scope != plugins.ProjectScope {
			var extensionSpecEnvVars []models.ExtensionSpecEnvironmentVariable
			if r.db.Where("environment_variable_id = ?", args.EnvironmentVariable.ID).Find(&extensionSpecEnvVars).RecordNotFound() {
				log.InfoWithFields("Nothing to update", log.Fields{
					"envVar": existingEnvVar,
				})
			}
			for _, extensionSpecEnvVar := range extensionSpecEnvVars {
				r.db.Delete(&extensionSpecEnvVar)
			}
		}
		r.actions.EnvironmentVariableDeleted(&existingEnvVar)
		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: existingEnvVar}, nil
	}
}

func (r *EnvironmentVariableResolver) ID() graphql.ID {
	return graphql.ID(r.EnvironmentVariable.Model.ID.String())
}

func (r *EnvironmentVariableResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project
	r.db.Model(r.EnvironmentVariable).Related(&project)
	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *EnvironmentVariableResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var env models.Environment
	r.db.Model(r.EnvironmentVariable).Related(&env)
	return &EnvironmentResolver{db: r.db, Environment: env}, nil
}

func (r *EnvironmentVariableResolver) Key() string {
	return r.EnvironmentVariable.Key
}

func (r *EnvironmentVariableResolver) Value() string {
	return r.EnvironmentVariable.Value
}

func (r *EnvironmentVariableResolver) Version() int32 {
	return r.EnvironmentVariable.Version
}

func (r *EnvironmentVariableResolver) Type() string {
	return string(r.EnvironmentVariable.Type)
}

func (r *EnvironmentVariableResolver) Scope() string {
	return string(r.EnvironmentVariable.Scope)
}

func (r *EnvironmentVariableResolver) User() (*UserResolver, error) {
	var user models.User
	r.db.Model(r.EnvironmentVariable).Related(&user)
	return &UserResolver{db: r.db, User: user}, nil
}

func (r *EnvironmentVariableResolver) Versions(ctx context.Context) ([]*EnvironmentVariableResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}
	var rows []models.EnvironmentVariable
	var results []*EnvironmentVariableResolver

	spew.Dump(r.EnvironmentVariable)

	r.db.Unscoped().Where("project_id = ? and key = ? and environment_id = ?", r.EnvironmentVariable.ProjectId, r.EnvironmentVariable.Key, r.EnvironmentVariable.EnvironmentId).Order("version desc").Find(&rows)

	for _, envVar := range rows {
		results = append(results, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar})
	}

	return results, nil
}
