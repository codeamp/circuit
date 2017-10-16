package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionSpecInput struct {
	ID        *string
	Name      *string
	Component *string
	FormSpec  *string
	Type      *string
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
	extensionSpec := models.ExtensionSpec{
		Name:      *args.ExtensionSpec.Name,
		Component: *args.ExtensionSpec.Component,
		FormSpec:  *args.ExtensionSpec.FormSpec,
		Type:      *args.ExtensionSpec.Type,
	}

	r.db.Create(&extensionSpec)
	r.actions.ExtensionSpecCreated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) UpdateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	var extensionSpec models.ExtensionSpec

	if r.db.Where("id = ?", args.ExtensionSpec.ID).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ExtensionSpec,
		})
		return &ExtensionSpecResolver{}, nil
	}

	extensionSpec.Name = *args.ExtensionSpec.Name
	extensionSpec.Component = *args.ExtensionSpec.Component
	extensionSpec.FormSpec = *args.ExtensionSpec.FormSpec
	extensionSpec.Type = *args.ExtensionSpec.Type

	r.db.Save(&extensionSpec)
	r.actions.ExtensionSpecUpdated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) DeleteExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	extensionSpec := models.ExtensionSpec{}

	extensionSpecId, err := uuid.FromString(*args.ExtensionSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.db.Where("id=?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
		return nil, fmt.Errorf("ExtensionSpec not found with given argument id")
	}

	r.db.Delete(extensionSpec)

	r.actions.ExtensionSpecDeleted(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
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
	return r.ExtensionSpec.Type
}

func (r *ExtensionSpecResolver) FormSpec() string {
	return r.ExtensionSpec.FormSpec
}

func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Created}
}
