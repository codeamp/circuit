package resolvers

import (
	"github.com/jinzhu/gorm/dialects/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/scalar"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
}

func NewExtensionResolver(extension models.Extension, db *gorm.DB) *ExtensionResolver {
	return &ExtensionResolver{
		db:        db,
		Extension: extension,
	}
}

type ExtensionInput struct {
	ID              *string
	ProjectId       string
	ExtensionSpecId string
	Config          scalar.Json
	EnvironmentId   string
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

func (r *ExtensionResolver) Config(ctx context.Context) scalar.Json {
	return scalar.Json{r.Extension.Config.RawMessage}
}

func (r *ExtensionResolver) Artifacts() scalar.Json {
	return scalar.Json{r.Extension.Artifacts.RawMessage}
}

func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Extension.Model.CreatedAt}
}

func (r *Resolver) CreateExtension(ctx context.Context, args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	extensionSpecId, err := uuid.FromString(args.Extension.ExtensionSpecId)
	if err != nil {
		log.InfoWithFields("couldn't parse ExtensionSpecId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse ExtensionSpecId. Invalid Format.")
	}

	projectId, err := uuid.FromString(args.Extension.ProjectId)
	if err != nil {
		log.InfoWithFields("couldn't parse ProjectId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse ProjectId. Invalid format.")
	}

	environmentId, err := uuid.FromString(args.Extension.EnvironmentId)
	if err != nil {
		log.InfoWithFields("couldn't parse EnvironmentId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse EnvironmentId. Invalid format.")
	}

	// check if extension already exists with project
	if r.db.Where("project_id = ? and extension_spec_id = ? and environment_id = ?", projectId, extensionSpecId, environmentId).Find(&extension).RecordNotFound() {
		// make sure extension form spec values are valid
		// if they are valid, create extension object
		var extensionSpec models.ExtensionSpec
		if r.db.Where("id = ?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("can't find corresponding extensionSpec", log.Fields{
				"extension": args.Extension,
			})
			return nil, errors.New("Can't find corresponding extensionSpec.")
		}

		extension = models.Extension{
			ExtensionSpecId: extensionSpecId,
			ProjectId:       projectId,
			EnvironmentId:   environmentId,
			Config:          postgres.Jsonb{[]byte(args.Extension.Config.RawMessage)},
			State:           plugins.GetState("waiting"),
			Artifacts:       postgres.Jsonb{},
		}
		r.db.Save(&extension)

		go r.actions.ExtensionCreated(&extension)
		return &ExtensionResolver{db: r.db, Extension: extension}, nil
	}
	return nil, errors.New("This extension is already installed in this project.")
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension

	if r.db.Where("id = ?", args.Extension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}
	extension.Config = postgres.Jsonb{args.Extension.Config.RawMessage}
	extension.State = plugins.GetState("waiting")

	r.db.Save(&extension)
	r.actions.ExtensionUpdated(&extension)

	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *Resolver) DeleteExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	var res []models.ReleaseExtension

	if r.db.Where("id = ?", args.Extension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}

	// delete all release extension objects with extension id
	if r.db.Where("extension_id = ?", args.Extension.ID).Find(&res).RecordNotFound() {
		log.InfoWithFields("no release extensions found", log.Fields{
			"extension": extension,
		})
		return &ExtensionResolver{}, nil
	}

	for _, re := range res {
		r.db.Delete(&re)
	}

	r.db.Delete(&extension)
	r.actions.ExtensionDeleted(&extension)
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ExtensionResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment models.Environment
	if r.db.Where("id = ?", r.Extension.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.Extension.EnvironmentId,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}
