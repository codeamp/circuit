package graphql_resolver

import (
	"context"
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
)

// ReleaseResolver resolver for Release
type ReleaseResolver struct {
	DBReleaseResolver *db_resolver.ReleaseResolver
}

// ID
func (r *ReleaseResolver) ID() graphql.ID {
	return graphql.ID(r.DBReleaseResolver.Release.Model.ID.String())
}

// Project
func (r *ReleaseResolver) Project() *ProjectResolver {
	// return r.DBReleaseResolver.Project()
	return nil
}

// User
func (r *ReleaseResolver) User() *UserResolver {
	// return r.DBReleaseResolver.User()
	return nil
}

// Artifacts
func (r *ReleaseResolver) Artifacts(ctx context.Context) (model.JSON, error) {
	// return r.DBReleaseResolver.Artifacts(ctx)
	return model.JSON{}, nil
}

// HeadFeature
func (r *ReleaseResolver) HeadFeature() *FeatureResolver {
	// return r.DBReleaseResolver.HeadFeature()
	return nil
}

// TailFeature
func (r *ReleaseResolver) TailFeature() *FeatureResolver {
	// return r.DBReleaseResolver.TailFeature()
	return nil
}

// State
func (r *ReleaseResolver) State() string {
	return string(r.DBReleaseResolver.Release.State)
}

// ReleaseExtensions
func (r *ReleaseResolver) ReleaseExtensions() []*ReleaseExtensionResolver {
	// return r.DBReleaseResolver.ReleaseExtensions()
	return nil
}

// StateMessage
func (r *ReleaseResolver) StateMessage() string {
	return r.DBReleaseResolver.Release.StateMessage
}

// Environment
func (r *ReleaseResolver) Environment() (*EnvironmentResolver, error) {
	// return r.DBReleaseResolver.Environment()
	return nil, nil
}

// Created
func (r *ReleaseResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBReleaseResolver.Release.Model.CreatedAt}
}

func (r *ReleaseResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBReleaseResolver.Release)
}

func (r *ReleaseResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBReleaseResolver.Release)
}
