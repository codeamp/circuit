package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type EnvironmentInput struct {
	ID        *string
	Name      string
	GitBranch *string
}

type EnvironmentResolver struct {
	db          *gorm.DB
	Environment models.Environment
}

func (r *Resolver) Environment(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentResolver, error) {
	env := models.Environment{}
	if err := r.db.Where("id = ?", args.ID).First(&env).Error; err != nil {
		return nil, err
	}

	return &EnvironmentResolver{db: r.db, Environment: env}, nil
}

func (r *Resolver) CreateEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {

	var existingEnv models.Environment

	branch := "master"
	if args.Environment.GitBranch != nil {
		branch = *args.Environment.GitBranch
	}

	if r.db.Where("name = ?", args.Environment.Name).Find(&existingEnv).RecordNotFound() {
		env := models.Environment{
			Name:      args.Environment.Name,
		}

		r.db.Create(&env)

		r.actions.EnvironmentCreated(&env)

		return &EnvironmentResolver{db: r.db, Environment: env}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironment: name already exists")
	}
}

func (r *Resolver) UpdateEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv models.Environment
	if r.db.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		existingEnv.Name = args.Environment.Name

		r.db.Save(&existingEnv)
		r.actions.EnvironmentUpdated(&existingEnv)
		return &EnvironmentResolver{db: r.db, Environment: existingEnv}, nil
	}
}

func (r *Resolver) DeleteEnvironment(ctx context.Context, args *struct{ Environment *EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv models.Environment
	if r.db.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		existingEnv.Name = args.Environment.Name
		r.db.Delete(&existingEnv)
		r.actions.EnvironmentDeleted(&existingEnv)
		return &EnvironmentResolver{db: r.db, Environment: existingEnv}, nil
	}
}

func (r *EnvironmentResolver) ID() graphql.ID {
	return graphql.ID(r.Environment.Model.ID.String())
}

func (r *EnvironmentResolver) Name(ctx context.Context) string {
	return r.Environment.Name
}

func (r *EnvironmentResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Environment.Model.CreatedAt}
}
