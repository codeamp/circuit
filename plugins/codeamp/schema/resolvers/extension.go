package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ExtensionInput struct {
	ID        *string
	Name      string
	Component string
	Hooks     []string
	FormSpec  string
	Type      string
}

func (r *Resolver) Extension(ctx context.Context, args *struct{ ID graphql.ID }) *ExtensionResolver {
	extension := models.Extension{}
	return &ExtensionResolver{db: r.db, Extension: extension}
}

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
}

func (r *Resolver) CreateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	extension := models.Extension{
		Name:      args.Extension.Name,
		Component: args.Extension.Component,
		FormSpec:  args.Extension.FormSpec,
		Type:      args.Extension.Type,
	}

	r.db.Create(&extension)
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension

	if r.db.Where("id = ?", args.Extension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}

	extension.Name = args.Extension.Name
	extension.Component = args.Extension.Component
	extension.FormSpec = args.Extension.FormSpec
	extension.Type = args.Extension.Type

	r.db.Save(&extension)
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.Extension.Model.ID.String())
}

func (r *ExtensionResolver) Name() string {
	return r.Extension.Name
}

func (r *ExtensionResolver) Component() string {
	return r.Extension.Component
}

func (r *ExtensionResolver) Type() string {
	return r.Extension.Type
}

func (r *ExtensionResolver) FormSpec() string {
	return r.Extension.FormSpec
}

func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Extension.Created}
}
