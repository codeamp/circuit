package resolvers

import (
	"github.com/satori/go.uuid"
	"context"
	"fmt"
	log "github.com/codeamp/logger"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type EnvironmentProjectBranchInput struct {
	ID        *string
	EnvironmentId string
	ProjectId string
	GitBranch string
}

type EnvironmentProjectBranchResolver struct {
	db          *gorm.DB
	EnvironmentProjectBranch models.EnvironmentProjectBranch
}

func (r *Resolver) EnvironmentProjectBranch(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentProjectBranchResolver, error) {
	env := models.EnvironmentProjectBranch{}
	if err := r.db.Where("id = ?", args.ID).First(&env).Error; err != nil {
		return nil, err
	}

	return &EnvironmentProjectBranchResolver{db: r.db, EnvironmentProjectBranch: env}, nil
}

func (r *Resolver) CreateEnvironmentProjectBranch(ctx context.Context, args *struct{ EnvironmentProjectBranch *EnvironmentProjectBranchInput }) (*EnvironmentProjectBranchResolver, error) {

	var existingEnvProjectBranch models.EnvironmentProjectBranch

	if r.db.Where("name = ?", args.EnvironmentProjectBranch.GitBranch).Find(&existingEnvProjectBranch).RecordNotFound() {
		environment := models.Environment{}
		project := models.Project{}

		if r.db.Where("id = ?", args.EnvironmentProjectBranch.EnvironmentId).Find(&environment).RecordNotFound() {
			log.InfoWithFields("environment does not exist", log.Fields{
				"id": args.EnvironmentProjectBranch.EnvironmentId,
			})
			return nil, fmt.Errorf("Environment does not exist.")
		}
		if r.db.Where("id = ?", args.EnvironmentProjectBranch.ProjectId).Find(&project).RecordNotFound() {
			log.InfoWithFields("project does not exist", log.Fields{
				"id": args.EnvironmentProjectBranch.ProjectId,
			})
			return nil, fmt.Errorf("Project does not exist.")
		}

		env := models.EnvironmentProjectBranch{
			GitBranch: args.EnvironmentProjectBranch.GitBranch,
			EnvironmentId: environment.Model.ID,
			ProjectId: project.Model.ID,
		}

		r.db.Create(&env)

		r.actions.EnvironmentProjectBranchCreated(&env)

		return &EnvironmentProjectBranchResolver{db: r.db, EnvironmentProjectBranch: env}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironmentProjectBranch: name already exists")
	}
}

func (r *Resolver) UpdateEnvironmentProjectBranch(ctx context.Context, args *struct{ EnvironmentProjectBranch *EnvironmentProjectBranchInput }) (*EnvironmentProjectBranchResolver, error) {
	var existingEnvProjectBranch models.EnvironmentProjectBranch

	if r.db.Where("id = ?", args.EnvironmentProjectBranch.ID).Find(&existingEnvProjectBranch).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnv: couldn't find EnvironmentProjectBranch: %s", *args.EnvironmentProjectBranch.ID)
	} else {
		existingEnvProjectBranch.GitBranch =  args.EnvironmentProjectBranch.GitBranch
		r.db.Save(&existingEnvProjectBranch)
		r.actions.EnvironmentProjectBranchUpdated(&existingEnvProjectBranch)
		return &EnvironmentProjectBranchResolver{db: r.db, EnvironmentProjectBranch: existingEnvProjectBranch}, nil
	}
}

func (r *Resolver) DeleteEnvironmentProjectBranch(ctx context.Context, args *struct{ EnvironmentProjectBranch *EnvironmentProjectBranchInput }) (*EnvironmentProjectBranchResolver, error) {
	var existingEnvProjectBranch models.EnvironmentProjectBranch
	if r.db.Where("id = ?", args.EnvironmentProjectBranch.ID).Find(&existingEnvProjectBranch).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find EnvironmentProjectBranch: %s", *args.EnvironmentProjectBranch.ID)
	} else {
		existingEnvProjectBranch.GitBranch = args.EnvironmentProjectBranch.GitBranch
		r.db.Delete(&existingEnvProjectBranch)
		r.actions.EnvironmentProjectBranchDeleted(&existingEnvProjectBranch)
		return &EnvironmentProjectBranchResolver{db: r.db, EnvironmentProjectBranch: existingEnvProjectBranch}, nil
	}
}

func (r *EnvironmentProjectBranchResolver) ID() graphql.ID {
	return graphql.ID(r.EnvironmentProjectBranch.Model.ID.String())
}

func (r *EnvironmentProjectBranchResolver) GitBranch(ctx context.Context) string {
	return r.EnvironmentProjectBranch.GitBranch
}

func (r *EnvironmentProjectBranchResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentProjectBranch.Model.CreatedAt}
}
