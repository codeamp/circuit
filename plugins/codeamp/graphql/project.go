package graphql_resolver

import (
	"context"
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// ProjectResolver resolver for Project
type ProjectResolver struct {
	DBProjectResolver *db_resolver.ProjectResolver
}

// ID
func (r *ProjectResolver) ID() graphql.ID {
	return graphql.ID(r.DBProjectResolver.Project.Model.ID.String())
}

// Name
func (r *ProjectResolver) Name() string {
	return r.DBProjectResolver.Project.Name
}

// Slug
func (r *ProjectResolver) Slug() string {
	return r.DBProjectResolver.Project.Slug
}

// Repository
func (r *ProjectResolver) Repository() string {
	return r.DBProjectResolver.Project.Repository
}

// Secret
func (r *ProjectResolver) Secret() string {
	return r.DBProjectResolver.Project.Secret
}

// GitUrl
func (r *ProjectResolver) GitUrl() string {
	return r.DBProjectResolver.Project.GitUrl
}

// GitProtocol
func (r *ProjectResolver) GitProtocol() string {
	return r.DBProjectResolver.Project.GitProtocol
}

// RsaPrivateKey
func (r *ProjectResolver) RsaPrivateKey() string {
	return r.DBProjectResolver.Project.RsaPrivateKey
}

// RsaPublicKey
func (r *ProjectResolver) RsaPublicKey() string {
	return r.DBProjectResolver.Project.RsaPublicKey
}

// Features
func (r *ProjectResolver) Features(showDeployed bool) []*FeatureResolver {
	// return r.DBProjectResolver.Features(showDeployed)
	return nil
}

// CurrentRelease
func (r *ProjectResolver) CurrentRelease() (*ReleaseResolver, error) {
	// return r.DBProjectResolver.CurrentRelease()
	return nil, nil
}

// Releases
func (r *ProjectResolver) Releases() []*ReleaseResolver {
	// return r.DBProjectResolver.Releases()
	return nil
}

// Services
func (r *ProjectResolver) Services() []*ServiceResolver {
	// return r.DBProjectResolver.Services()
	return nil
}

// Secrets
func (r *ProjectResolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	// return r.DBProjectResolver.Secrets(ctx)
	return nil, nil
}

// ProjectExtensions
func (r *ProjectResolver) Extensions() ([]*ProjectExtensionResolver, error) {
	// return r.DBProjectResolver.Extensions()
	return nil, nil
}

// GitBranch
func (r *ProjectResolver) GitBranch() string {
	// return r.DBProjectResolver.GitBranch()
	return ""
}

// ContinuousDeploy
func (r *ProjectResolver) ContinuousDeploy() bool {
	// return r.DBProjectResolver.ContinuousDeploy()
	return false
}

// Environments
func (r *ProjectResolver) Environments() []*EnvironmentResolver {
	// return r.DBProjectResolver.Environments()
	return nil
}

// Bookmarked
func (r *ProjectResolver) Bookmarked(ctx context.Context) bool {
	// return r.DBProjectResolver.Bookmarked()
	return false
}

// Created
func (r *ProjectResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBProjectResolver.Project.Model.CreatedAt}
}

func (r *ProjectResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBProjectResolver.Project)
}

func (r *ProjectResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBProjectResolver.Project)
}
