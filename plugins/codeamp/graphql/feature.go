package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// FeatureResolver resolver for Feature
type FeatureResolver struct {
	DBFeatureResolver *db_resolver.FeatureResolver
}

// ID
func (r *FeatureResolver) ID() graphql.ID {
	return graphql.ID(r.DBFeatureResolver.Feature.Model.ID.String())
}

// Project
func (r *FeatureResolver) Project() *ProjectResolver {
	return &ProjectResolver{DBProjectResolver: r.DBFeatureResolver.Project()}
}

// Message
func (r *FeatureResolver) Message() string {
	return r.DBFeatureResolver.Feature.Message
}

// User
func (r *FeatureResolver) User() string {
	return r.DBFeatureResolver.Feature.User
}

// Hash
func (r *FeatureResolver) Hash() string {
	return r.DBFeatureResolver.Feature.Hash
}

// ParentHash
func (r *FeatureResolver) ParentHash() string {
	return r.DBFeatureResolver.Feature.ParentHash
}

// Ref
func (r *FeatureResolver) Ref() string {
	return r.DBFeatureResolver.Feature.Ref
}

// Created
func (r *FeatureResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBFeatureResolver.Feature.Created}
}

// NotFoundSince
func (r *FeatureResolver) NotFoundSince() graphql.Time {
	return graphql.Time{Time: r.DBFeatureResolver.Feature.NotFoundSince}
}

func (r *FeatureResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBFeatureResolver.Feature)
}

func (r *FeatureResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBFeatureResolver.Feature)
}
