package resolvers

import (
	"context"
	"fmt"
	log "github.com/codeamp/logger"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type EnvironmentBasedProjectBranchInput struct {
	ID        *string
	EnvironmentId string
	ProjectId string
	GitBranch string
}

type EnvironmentBasedProjectBranchResolver struct {
	db          *gorm.DB
	EnvironmentBasedProjectBranch models.EnvironmentBasedProjectBranch
}

func (r *Resolver) EnvironmentBasedProjectBranch(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentBasedProjectBranchResolver, error) {
	env := models.EnvironmentBasedProjectBranch{}
	if err := r.db.Where("id = ?", args.ID).First(&env).Error; err != nil {
		return nil, err
	}

	return &EnvironmentBasedProjectBranchResolver{db: r.db, EnvironmentBasedProjectBranch: env}, nil
}

func (r *Resolver) CreateEnvironmentBasedProjectBranch(ctx context.Context, args *struct{ EnvironmentBasedProjectBranch *EnvironmentBasedProjectBranchInput }) (*EnvironmentBasedProjectBranchResolver, error) {

	var existingEnvProjectBranch models.EnvironmentBasedProjectBranch

	if r.db.Where("git_branch = ? and project_id = ? and environment_id = ?", args.EnvironmentBasedProjectBranch.GitBranch, 
		args.EnvironmentBasedProjectBranch.ProjectId, args.EnvironmentBasedProjectBranch.EnvironmentId).Find(&existingEnvProjectBranch).RecordNotFound() {
		environment := models.Environment{}
		project := models.Project{}

		if r.db.Where("id = ?", args.EnvironmentBasedProjectBranch.EnvironmentId).Find(&environment).RecordNotFound() {
			log.InfoWithFields("environment does not exist", log.Fields{
				"id": args.EnvironmentBasedProjectBranch.EnvironmentId,
			})
			return nil, fmt.Errorf("Environment does not exist.")
		}
		if r.db.Where("id = ?", args.EnvironmentBasedProjectBranch.ProjectId).Find(&project).RecordNotFound() {
			log.InfoWithFields("project does not exist", log.Fields{
				"id": args.EnvironmentBasedProjectBranch.ProjectId,
			})
			return nil, fmt.Errorf("Project does not exist.")
		}

		env := models.EnvironmentBasedProjectBranch{
			GitBranch: args.EnvironmentBasedProjectBranch.GitBranch,
			EnvironmentId: environment.Model.ID,
			ProjectId: project.Model.ID,
		}

		r.db.Create(&env)

		r.actions.EnvironmentBasedProjectBranchCreated(&env)

		return &EnvironmentBasedProjectBranchResolver{db: r.db, EnvironmentBasedProjectBranch: env}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironmentBasedProjectBranch: gitBranch already exists")
	}
}

func (r *Resolver) UpdateEnvironmentBasedProjectBranch(ctx context.Context, args *struct{ EnvironmentBasedProjectBranch *EnvironmentBasedProjectBranchInput }) (*EnvironmentBasedProjectBranchResolver, error) {
	var existingEnvProjectBranch models.EnvironmentBasedProjectBranch

	if r.db.Where("id = ?", args.EnvironmentBasedProjectBranch.ID).Find(&existingEnvProjectBranch).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnv: couldn't find EnvironmentBasedProjectBranch: %s", *args.EnvironmentBasedProjectBranch.ID)
	} else {
		existingEnvProjectBranch.GitBranch =  args.EnvironmentBasedProjectBranch.GitBranch
		r.db.Save(&existingEnvProjectBranch)
		r.actions.EnvironmentBasedProjectBranchUpdated(&existingEnvProjectBranch)
		return &EnvironmentBasedProjectBranchResolver{db: r.db, EnvironmentBasedProjectBranch: existingEnvProjectBranch}, nil
	}
}

func (r *Resolver) DeleteEnvironmentBasedProjectBranch(ctx context.Context, args *struct{ EnvironmentBasedProjectBranch *EnvironmentBasedProjectBranchInput }) (*EnvironmentBasedProjectBranchResolver, error) {
	var existingEnvProjectBranch models.EnvironmentBasedProjectBranch
	if r.db.Where("id = ?", args.EnvironmentBasedProjectBranch.ID).Find(&existingEnvProjectBranch).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find EnvironmentBasedProjectBranch: %s", *args.EnvironmentBasedProjectBranch.ID)
	} else {
		existingEnvProjectBranch.GitBranch = args.EnvironmentBasedProjectBranch.GitBranch
		r.db.Delete(&existingEnvProjectBranch)
		r.actions.EnvironmentBasedProjectBranchDeleted(&existingEnvProjectBranch)
		return &EnvironmentBasedProjectBranchResolver{db: r.db, EnvironmentBasedProjectBranch: existingEnvProjectBranch}, nil
	}
}

func (r *EnvironmentBasedProjectBranchResolver) ID() graphql.ID {
	return graphql.ID(r.EnvironmentBasedProjectBranch.Model.ID.String())
}

func (r *EnvironmentBasedProjectBranchResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment models.Environment
	if r.db.Where("id = ?", r.EnvironmentBasedProjectBranch.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.EnvironmentBasedProjectBranch.EnvironmentId,
		})
		return nil, fmt.Errorf("Environment not found.")		
	}
	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}

func (r *EnvironmentBasedProjectBranchResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project
	if r.db.Where("id = ?", r.EnvironmentBasedProjectBranch.ProjectId).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": r.EnvironmentBasedProjectBranch.ProjectId,
		})
		return nil, fmt.Errorf("Project not found.")		
	}
	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *EnvironmentBasedProjectBranchResolver) GitBranch(ctx context.Context) string {
	return r.EnvironmentBasedProjectBranch.GitBranch
}

func (r *EnvironmentBasedProjectBranchResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentBasedProjectBranch.Model.CreatedAt}
}
