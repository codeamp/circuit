package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

func (r *Resolver) Release(ctx context.Context, args *struct{ ID graphql.ID }) *ReleaseResolver {
	release := models.Release{}
	return &ReleaseResolver{db: r.db, Release: release}
}

type ReleaseResolver struct {
	db      *gorm.DB
	Release models.Release
}

type ReleaseInput struct {
	ID            *string
	ProjectId     string
	HeadFeatureId string
	EnvironmentId string
}

func (r *Resolver) CreateRelease(ctx context.Context, args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	fmt.Println("CreateRelease")

	var tailFeatureId uuid.UUID
	var currentRelease models.Release

	projectId, err := uuid.FromString(args.Release.ProjectId)
	if err != nil {
		log.InfoWithFields("Couldn't parse projectId", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Couldn't parse projectId")
	}
	headFeatureId, err := uuid.FromString(args.Release.HeadFeatureId)
	if err != nil {
		log.InfoWithFields("Couldn't parse headFeatureId", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Couldn't parse headFeatureId")
	}
	environmentId, err := uuid.FromString(args.Release.EnvironmentId)
	if err != nil {
		log.InfoWithFields("Couldn't parse environmentId", log.Fields{
			"args": args,
		})
		return nil, fmt.Errorf("Couldn't parse environmentId")
	}

	// the tail feature id is the current release's head feature id
	if r.db.Where("state = ? and project_id = ? and environment_id = ?", plugins.Complete, args.Release.ProjectId, environmentId).Find(&currentRelease).Order("created_at desc").Limit(1).RecordNotFound() {
		// get first ever feature in project if current release doesn't exist yet
		var firstFeature models.Feature
		if r.db.Where("project_id = ?", args.Release.ProjectId).Find(&firstFeature).Order("created_at asc").Limit(1).RecordNotFound() {
			log.InfoWithFields("CreateRelease", log.Fields{
				"release": r,
			})
			return nil, fmt.Errorf("No features found.")
		}
		tailFeatureId = firstFeature.ID
	} else {
		tailFeatureId = currentRelease.HeadFeatureID
	}

	userIdString, err := utils.CheckAuth(ctx, []string{})
	if err != nil {
		return &ReleaseResolver{}, err
	}

	userId := uuid.FromStringOrNil(userIdString)

	release := models.Release{
		ProjectId:     projectId,
		EnvironmentId: environmentId,
		UserID:        userId,
		HeadFeatureID: headFeatureId,
		TailFeatureID: tailFeatureId,
		State:         plugins.Waiting,
		StateMessage:  "Release created",
	}

	r.db.Create(&release)
	r.actions.ReleaseCreated(&release)

	return &ReleaseResolver{db: r.db, Release: release}, nil
}

func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.Release.Model.ID.String())
}

func (r *ReleaseResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project

	r.db.Model(r.Release).Related(&project)

	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ReleaseResolver) User(ctx context.Context) (*UserResolver, error) {
	var user models.User

	r.db.Model(r.User).Related(&user)

	return &UserResolver{db: r.db, User: user}, nil
}

func (r *ReleaseResolver) HeadFeature() (*FeatureResolver, error) {
	var feature models.Feature

	r.db.Where("id = ?", r.Release.HeadFeatureID).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	var rows []models.ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.db.Where("release_id = ?", r.Release.ID).Find(&rows)
	for _, re := range rows {
		results = append(results, &ReleaseExtensionResolver{db: r.db, ReleaseExtension: re})
	}
	return results, nil

}

func (r *ReleaseResolver) TailFeature() (*FeatureResolver, error) {
	var feature models.Feature

	r.db.Where("id = ?", r.Release.TailFeatureID).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) State() string {
	return string(r.Release.State)
}

func (r *ReleaseResolver) StateMessage() string {
	return r.Release.StateMessage
}

func (r *ReleaseResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment models.Environment
	if r.db.Where("id = ?", r.Release.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"service": r.Release,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}
