package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
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
