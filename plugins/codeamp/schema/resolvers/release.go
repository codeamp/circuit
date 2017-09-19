package codeamp_schema_resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

func (r *Resolver) Release(ctx context.Context, args *struct{ ID graphql.ID }) *ReleaseResolver {
	release := codeamp_models.Release{}
	return &ReleaseResolver{db: r.db, Release: release}
}

type ReleaseResolver struct {
	db      *gorm.DB
	Release codeamp_models.Release
}

type ReleaseInput struct {
	FeatureId string
}

func (r *ReleaseResolver) CreateRelease(args *struct{ Release *ReleaseInput }) (*ReleaseResolver, error) {
	fmt.Println("CreateRelease")
	spew.Dump(*args.Release)
	return &ReleaseResolver{}, nil
}

func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.Release.Model.ID.String())
}

func (r *ReleaseResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project codeamp_models.Project

	r.db.Model(r.Release).Related(&project)

	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ReleaseResolver) User(ctx context.Context) (*UserResolver, error) {
	var user codeamp_models.User

	r.db.Model(r.User).Related(&user)

	return &UserResolver{db: r.db, User: user}, nil
}

func (r *ReleaseResolver) HeadFeature() (*FeatureResolver, error) {
	var feature codeamp_models.Feature

	r.db.Where("id = ?", r.Release.HeadFeatureId).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) TailFeature() (*FeatureResolver, error) {
	var feature codeamp_models.Feature

	r.db.Where("id = ?", r.Release.TailFeatureId).First(&feature)

	return &FeatureResolver{db: r.db, Feature: feature}, nil
}

func (r *ReleaseResolver) State() string {
	return string(r.Release.State)
}

func (r *ReleaseResolver) StateMessage() string {
	return r.Release.StateMessage
}
