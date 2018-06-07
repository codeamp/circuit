package graphql_resolver

import (
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// FeatureResolver resolver for Feature
type FeatureResolver struct {
	model.Feature
	DB *gorm.DB
}

// ID
func (r *FeatureResolver) ID() graphql.ID {
	return graphql.ID(r.Feature.Model.ID.String())
}

// Project
func (r *FeatureResolver) Project() *ProjectResolver {
	var project model.Project

	r.DB.Model(r.Feature).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// Message
func (r *FeatureResolver) Message() string {
	return r.Feature.Message
}

// User
func (r *FeatureResolver) User() string {
	return r.Feature.User
}

// Hash
func (r *FeatureResolver) Hash() string {
	return r.Feature.Hash
}

// ParentHash
func (r *FeatureResolver) ParentHash() string {
	return r.Feature.ParentHash
}

// Ref
func (r *FeatureResolver) Ref() string {
	return r.Feature.Ref
}

// Created
func (r *FeatureResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Feature.Created}
}

func (r *FeatureResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.Feature)
}

func (r *FeatureResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.Feature)
}
