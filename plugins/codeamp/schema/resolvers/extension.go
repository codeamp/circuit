package resolvers

import (
	"context"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
}

type ExtensionInput struct {
	ID              *string
	ProjectId       string
	ExtensionSpecId string
	FormSpecValues  string
}

func (r *Resolver) Extension(ctx context.Context, args *struct{ ID graphql.ID }) (*ExtensionResolver, error) {
	extension := models.Extension{}
	if err := r.db.Where("id = ?", args.ID).First(&extension).Error; err != nil {
		return nil, err
	}

	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.Extension.Model.ID.String())
}

func (r *ExtensionResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project
	r.db.Model(r.Extension).Related(&project)
	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ExtensionResolver) ExtensionSpec(ctx context.Context) (*ExtensionSpecResolver, error) {
	var extensionSpec models.ExtensionSpec
	r.db.Model(r.Extension).Related(&extensionSpec)
	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *ExtensionResolver) State() string {
	return string(r.Extension.State)
}

func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Extension.Created}
}

func (r *ExtensionResolver) Artifacts() string {
	return r.Extension.Artifacts
}

func (r *ExtensionResolver) FormSpecValues() string {
	return r.Extension.FormSpecValues
}

func (r *Resolver) CreateExtension(ctx context.Context, args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension

	// check if extension already exists with project
	if r.db.Where("project_id = ? and extension_spec_id = ?", args.Extension.ProjectId, args.Extension.ExtensionSpecId).Find(&extension).RecordNotFound() {
		// make sure extension form spec values are valid
		// if they are valid, create extension object

		extensionSpecId, err := uuid.FromString(args.Extension.ExtensionSpecId)
		if err != nil {
			log.InfoWithFields("couldn't parse ExtensionSpecId", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		projectId, err := uuid.FromString(args.Extension.ProjectId)
		if err != nil {
			log.InfoWithFields("couldn't parse ProjectId", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		extension = models.Extension{
			ExtensionSpecId: extensionSpecId,
			ProjectId:       projectId,
			FormSpecValues:  args.Extension.FormSpecValues,
			Artifacts:       "",
			State:           plugins.Waiting,
			Created:         time.Now(),
		}

		r.db.Create(&extension)
		r.actions.ExtensionCreated(&extension)
	}
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}
