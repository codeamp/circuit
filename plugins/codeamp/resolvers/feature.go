package codeamp_resolvers

import (
	"encoding/json"
	"time"

	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// Feature
type Feature struct {
	Model `json:",inline"`
	// ProjectID
	ProjectID uuid.UUID `bson:"projectID" json:"projectID" gorm:"type:uuid"`
	// Message
	Message string `json:"message"`
	// User
	User string `json:"user"`
	// Hash
	Hash string `json:"hash"`
	// ParentHash
	ParentHash string `json:"parentHash"`
	// Ref
	Ref string `json:"ref"`
	// Created
	Created time.Time `json:"created"`
}

// FeatureResolver resolver for Feature
type FeatureResolver struct {
	Feature
	DB *gorm.DB
}

// ID
func (r *FeatureResolver) ID() graphql.ID {
	return graphql.ID(r.Feature.Model.ID.String())
}

// Project
func (r *FeatureResolver) Project() *ProjectResolver {
	var project Project

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
