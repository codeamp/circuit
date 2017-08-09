package codeamp_schema_resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

func (r *Resolver) Release(ctx context.Context, args *struct{ ID graphql.ID }) *ReleaseResolver {
	release := codeamp_models.Release{}
	return &ReleaseResolver{DB: r.DB, Release: release}
}

type ReleaseResolver struct {
	DB      *gorm.DB
	Release codeamp_models.Release
}

func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.Release.Model.ID.String())
}

func (r *ReleaseResolver) User(ctx context.Context) (*UserResolver, error) {
	var user codeamp_models.User

	r.DB.Model(r.User).Related(&user)

	return &UserResolver{DB: r.DB, User: user}, nil
}
