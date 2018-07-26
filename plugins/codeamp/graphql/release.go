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
	return &ProjectResolver{DBProjectResolver: r.DBReleaseResolver.Project()}
}

// User
func (r *ReleaseResolver) User() *UserResolver {
	return &UserResolver{DBUserResolver: r.DBReleaseResolver.User()}
}

// Artifacts
func (r *ReleaseResolver) Artifacts(ctx context.Context) (model.JSON, error) {
	return r.DBReleaseResolver.Artifacts(ctx)
}

// HeadFeature
func (r *ReleaseResolver) HeadFeature() *FeatureResolver {
	return &FeatureResolver{DBFeatureResolver: r.DBReleaseResolver.HeadFeature()}
}

// TailFeature
func (r *ReleaseResolver) TailFeature() *FeatureResolver {
	return &FeatureResolver{DBFeatureResolver: r.DBReleaseResolver.TailFeature()}
}

// State
func (r *ReleaseResolver) State() string {
	return string(r.DBReleaseResolver.Release.State)
}

// IsRollback
func (r *ReleaseResolver) IsRollback() bool {
	return r.DBReleaseResolver.Release.IsRollback
}

// ReleaseExtensions
func (r *ReleaseResolver) ReleaseExtensions() []*ReleaseExtensionResolver {
	db_resolvers := r.DBReleaseResolver.ReleaseExtensions()
	gql_resolvers := make([]*ReleaseExtensionResolver, 0, len(db_resolvers))

	for _, i := range db_resolvers {
		gql_resolvers = append(gql_resolvers, &ReleaseExtensionResolver{DBReleaseExtensionResolver: i})
	}

	return gql_resolvers
}

// StateMessage
func (r *ReleaseResolver) StateMessage() string {
	return r.DBReleaseResolver.Release.StateMessage
}

// Environment
func (r *ReleaseResolver) Environment() (*EnvironmentResolver, error) {
	resolver, err := r.DBReleaseResolver.Environment()
	return &EnvironmentResolver{DBEnvironmentResolver: resolver}, err
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
