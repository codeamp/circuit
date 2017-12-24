package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/scalar"
	log "github.com/codeamp/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionSpecInput struct {
	ID            *string
	Name          string
	Component     string
	Type          string
	Key           string
	EnvironmentId string
	Config        scalar.Json
}

func (r *Resolver) ExtensionSpec(ctx context.Context, args *struct{ ID graphql.ID }) *ExtensionSpecResolver {
	extensionSpec := models.ExtensionSpec{}
	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}
}

type ExtensionSpecResolver struct {
	db            *gorm.DB
	ExtensionSpec models.ExtensionSpec
}

func (r *Resolver) CreateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	environmentId, err := uuid.FromString(args.ExtensionSpec.EnvironmentId)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentId")
	}

	extensionSpec := models.ExtensionSpec{
		Name:          args.ExtensionSpec.Name,
		Component:     args.ExtensionSpec.Component,
		Type:          plugins.Type(args.ExtensionSpec.Type),
		Key:           args.ExtensionSpec.Key,
		EnvironmentId: environmentId,
		Config:        postgres.Jsonb{args.ExtensionSpec.Config.RawMessage},
	}

	r.db.Create(&extensionSpec)
	r.actions.ExtensionSpecCreated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) UpdateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	extensionSpec := models.ExtensionSpec{}
	if r.db.Where("id = ?", args.ExtensionSpec.ID).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("could not find extensionspec with id", log.Fields{
			"id": args.ExtensionSpec.ID,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, fmt.Errorf("could not find extensionspec with id")
	}

	environmentId, err := uuid.FromString(args.ExtensionSpec.EnvironmentId)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentId")
	}

	// update extensionspec properties
	extensionSpec.Name = args.ExtensionSpec.Name
	extensionSpec.Key = args.ExtensionSpec.Key
	extensionSpec.Type = plugins.Type(args.ExtensionSpec.Type)
	extensionSpec.Component = args.ExtensionSpec.Component
	extensionSpec.EnvironmentId = environmentId
	extensionSpec.Config = postgres.Jsonb{args.ExtensionSpec.Config.RawMessage}

	r.db.Save(&extensionSpec)

	r.actions.ExtensionSpecUpdated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) DeleteExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	extensionSpec := models.ExtensionSpec{}
	extensions := []models.Extension{}
	extensionSpecId, err := uuid.FromString(*args.ExtensionSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.db.Where("id=?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
		return nil, fmt.Errorf("ExtensionSpec not found with given argument id")
	}

	// delete all extensions using extension spec
	if r.db.Where("extension_spec_id = ?", extensionSpecId).Find(&extensions).RecordNotFound() {
		log.InfoWithFields("no extensions using this extension spec", log.Fields{
			"extension spec": extensionSpec,
		})
	}

	if len(extensions) > 0 {
		return nil, fmt.Errorf("You must delete all extensions using this extension spec in order to delete this extension spec.")
	} else {
		r.db.Delete(&extensionSpec)
		r.actions.ExtensionSpecDeleted(&extensionSpec)

		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
	}
}

func (r *ExtensionSpecResolver) ID() graphql.ID {
	return graphql.ID(r.ExtensionSpec.Model.ID.String())
}

func (r *ExtensionSpecResolver) Name() string {
	return r.ExtensionSpec.Name
}

func (r *ExtensionSpecResolver) Component() string {
	return r.ExtensionSpec.Component
}

func (r *ExtensionSpecResolver) Type() string {
	return string(r.ExtensionSpec.Type)
}

func (r *ExtensionSpecResolver) Key() string {
	return r.ExtensionSpec.Key
}

func (r *ExtensionSpecResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	environment := models.Environment{}

	if r.db.Where("id = ?", r.ExtensionSpec.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.ExtensionSpec.EnvironmentId,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}

func (r *ExtensionSpecResolver) Config(ctx context.Context) (scalar.Json, error) {
	return scalar.Json{r.ExtensionSpec.Config.RawMessage}, nil
}

func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Model.CreatedAt}
}
