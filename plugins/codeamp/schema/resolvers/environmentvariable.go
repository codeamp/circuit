package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/utils"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type EnvironmentVariableInput struct {
	ID        *string
	Key       string
	Value     string
	Type      *string
	ProjectId *string
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

	return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
}

func (r *Resolver) CreateEnvironmentVariable(ctx context.Context, args *struct{ EnvironmentVariable *EnvironmentVariableInput }) (*EnvironmentVariableResolver, error) {
	projectId, err := uuid.FromString(*args.EnvironmentVariable.ProjectId)
	if err != nil {
		return &EnvironmentVariableResolver{}, err
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
	if r.db.Where("key = ? and project_id = ?", args.EnvironmentVariable.Key, args.EnvironmentVariable.ProjectId).Find(&existingEnvVar).RecordNotFound() {
		spew.Dump(args.EnvironmentVariable)
		envVar := models.EnvironmentVariable{
			Key:       args.EnvironmentVariable.Key,
			Value:     args.EnvironmentVariable.Value,
			ProjectId: projectId,
			Version:   int32(0),
			Type:      *args.EnvironmentVariable.Type,
			UserId:    userId,
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
	if r.db.Where("id = ?", args.EnvironmentVariable.ID).Find(&existingEnvVar).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnvironmentVariable: key doesn't exist.")
	} else {
		spew.Dump("HELLO THERE UPDATE")
		spew.Dump(existingEnvVar)
		spew.Dump(args.EnvironmentVariable)
		envVar := models.EnvironmentVariable{
			Key:       args.EnvironmentVariable.Key,
			Value:     args.EnvironmentVariable.Value,
			ProjectId: existingEnvVar.ProjectId,
			Version:   existingEnvVar.Version + int32(1),
			Type:      existingEnvVar.Type,
			UserId:    existingEnvVar.UserId,
		}

		r.db.Create(&envVar)
		r.actions.EnvironmentVariableUpdated(&envVar)

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
	return r.EnvironmentVariable.Type
}

func (r *EnvironmentVariableResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentVariable.Created}
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

	r.db.Where("project_id = ? and key = ?", r.EnvironmentVariable.ProjectId, r.EnvironmentVariable.Key).Order("created desc").Find(&rows)
	for _, envVar := range rows {
		results = append(results, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar})
	}

	return results, nil
}
