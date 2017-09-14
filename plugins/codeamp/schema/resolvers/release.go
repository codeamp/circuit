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

func (r *ReleaseResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project codeamp_models.Project

	r.DB.Model(r.Release).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}, nil
}

func (r *ReleaseResolver) User(ctx context.Context) (*UserResolver, error) {
	var user codeamp_models.User

	r.DB.Model(r.User).Related(&user)

	return &UserResolver{DB: r.DB, User: user}, nil
}

func (r *ReleaseResolver) HeadFeature() (*FeatureResolver, error) {
	var feature codeamp_models.Feature

	r.DB.Where("id = ?", r.Release.HeadFeatureId).First(&feature)

	return &FeatureResolver{DB: r.DB, Feature: feature}, nil
}

func (r *ReleaseResolver) TailFeature() (*FeatureResolver, error) {
	var feature codeamp_models.Feature

	r.DB.Where("id = ?", r.Release.TailFeatureId).First(&feature)

	return &FeatureResolver{DB: r.DB, Feature: feature}, nil
}

func (r *ReleaseResolver) State() string {
	return string(r.Release.State)
}

func (r *ReleaseResolver) StateMessage() string {
	return r.Release.StateMessage
}
