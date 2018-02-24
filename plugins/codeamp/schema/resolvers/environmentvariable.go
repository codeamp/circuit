package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/utils"

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
	IsSecret      bool
}

type EnvironmentVariableResolver struct {
	db                       *gorm.DB
	EnvironmentVariable      models.EnvironmentVariable
	EnvironmentVariableValue models.EnvironmentVariableValue
}

func (r *Resolver) EnvironmentVariable(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentVariableResolver, error) {
	envVar := models.EnvironmentVariable{}
	envVarValue := models.EnvironmentVariableValue{}

	return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar, EnvironmentVariableValue: envVarValue}, nil
}

func (r *Resolver) CreateEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {

	projectId := uuid.UUID{}
	var environmentId uuid.UUID
	var environmentVariableScope models.EnvironmentVariableScope

	if args.EnvironmentVariable.ProjectId != nil {
		projectId = uuid.FromStringOrNil(*args.EnvironmentVariable.ProjectId)
	}

	environmentVariableScope = models.GetEnvironmentVariableScope(args.EnvironmentVariable.Scope)
	if environmentVariableScope == models.EnvironmentVariableScope("unknown") {
		return nil, fmt.Errorf("Invalid env var scope.")
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
			ProjectId:     projectId,
			Type:          plugins.GetType(args.EnvironmentVariable.Type),
			Scope:         environmentVariableScope,
			EnvironmentId: environmentId,
			IsSecret:      args.EnvironmentVariable.IsSecret,
		}
		r.db.Create(&envVar)

		envVarValue := models.EnvironmentVariableValue{
			EnvironmentVariableId: envVar.Model.ID,
			Value:  args.EnvironmentVariable.Value,
			UserId: userId,
		}
		r.db.Create(&envVarValue)

		r.actions.EnvironmentVariableCreated(&envVar)

		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar, EnvironmentVariableValue: envVarValue}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironmentVariable: key already exists")
	}

}

func (r *Resolver) UpdateEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {
	var envVar models.EnvironmentVariable

	userIdString, err := utils.CheckAuth(ctx, []string{})
	if err != nil {
		return &EnvironmentVariableResolver{}, err
	}

	userId, err := uuid.FromString(userIdString)
	if err != nil {
		return &EnvironmentVariableResolver{}, err
	}

	if r.db.Where("id = ?", args.EnvironmentVariable.ID).Find(&envVar).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnvironmentVariable: env var doesn't exist.")
	} else {
		envVarValue := models.EnvironmentVariableValue{
			EnvironmentVariableId: envVar.Model.ID,
			Value:  args.EnvironmentVariable.Value,
			UserId: userId,
		}
		r.db.Create(&envVarValue)
		r.actions.EnvironmentVariableUpdated(&envVar)

		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar, EnvironmentVariableValue: envVarValue}, nil
	}
}

func (r *Resolver) DeleteEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {
	var envVar models.EnvironmentVariable

	if r.db.Where("id = ?", args.EnvironmentVariable.ID).Find(&envVar).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnvironmentVariable: key doesn't exist.")
	} else {
		var rows []models.EnvironmentVariable

		r.db.Where("project_id = ? and key = ? and environment_id = ?", envVar.ProjectId, envVar.Key, envVar.EnvironmentId).Find(&rows)
		for _, ev := range rows {
			r.db.Unscoped().Delete(&ev)
		}

		r.actions.EnvironmentVariableDeleted(&envVar)

		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
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
	return r.EnvironmentVariableValue.Value
}

func (r *EnvironmentVariableResolver) Type() string {
	return string(r.EnvironmentVariable.Type)
}

func (r *EnvironmentVariableResolver) Scope() string {
	return string(r.EnvironmentVariable.Scope)
}

func (r *EnvironmentVariableResolver) IsSecret() bool {
	return r.EnvironmentVariable.IsSecret
}

func (r *EnvironmentVariableResolver) User() (*UserResolver, error) {
	var user models.User
	r.db.Model(r.EnvironmentVariableValue).Related(&user)
	return &UserResolver{db: r.db, User: user}, nil
}

func (r *EnvironmentVariableResolver) Versions(ctx context.Context) ([]*EnvironmentVariableResolver, error) {
	var envVarValues []models.EnvironmentVariableValue
	var envVarResolvers []*EnvironmentVariableResolver

	r.db.Where("environment_variable_id = ?", r.EnvironmentVariable.Model.ID).Order("created_at desc").Find(&envVarValues)

	for _, envVarValue := range envVarValues {
		envVarResolvers = append(envVarResolvers, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: r.EnvironmentVariable, EnvironmentVariableValue: envVarValue})
	}

	return envVarResolvers, nil
}

func (r *EnvironmentVariableResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentVariable.Model.CreatedAt}
}
