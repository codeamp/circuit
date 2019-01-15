package graphql_resolver

import (
	"context"
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
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
func (r *ProjectResolver) Features(args *struct {
	ShowDeployed *bool
	Params       *model.PaginatorInput
}) FeatureListResolver {
	db_resolver := r.DBProjectResolver.Features(args)
	return FeatureListResolver{
		DBFeatureListResolver: db_resolver,
	}
}

// CurrentRelease
func (r *ProjectResolver) CurrentRelease() (*ReleaseResolver, error) {
	resolver, err := r.DBProjectResolver.CurrentRelease()
	return &ReleaseResolver{DBReleaseResolver: resolver}, err
}

// Releases
func (r *ProjectResolver) Releases(args *struct {
	Params *model.PaginatorInput
}) *ReleaseListResolver {
	db_resolver := r.DBProjectResolver.Releases(args)

	return &ReleaseListResolver{
		DBReleaseListResolver: db_resolver,
	}
}

// Services
func (r *ProjectResolver) Services(args *struct {
	Params    *model.PaginatorInput
	SearchKey *string
}) *ServiceListResolver {
	db_resolver := r.DBProjectResolver.Services(args)

	return &ServiceListResolver{
		DBServiceListResolver: db_resolver,
	}
}

// Secrets
func (r *ProjectResolver) Secrets(ctx context.Context, args *struct {
	Params    *model.PaginatorInput
	SearchKey *string
}) (*SecretListResolver, error) {
	db_resolver, err := r.DBProjectResolver.Secrets(ctx, args)
	if err != nil {
		return nil, err
	}

	return &SecretListResolver{
		DBSecretListResolver: db_resolver,
	}, nil
}

// ProjectExtensions
func (r *ProjectResolver) Extensions() ([]*ProjectExtensionResolver, error) {
	db_resolvers, err := r.DBProjectResolver.Extensions()
	gql_resolvers := make([]*ProjectExtensionResolver, 0, len(db_resolvers))

	for _, i := range db_resolvers {
		gql_resolvers = append(gql_resolvers, &ProjectExtensionResolver{DBProjectExtensionResolver: i})
	}

	return gql_resolvers, err
}

// GitBranch
func (r *ProjectResolver) GitBranch() string {
	return r.DBProjectResolver.GitBranch()
}

// ContinuousDeploy
func (r *ProjectResolver) ContinuousDeploy() bool {
	return r.DBProjectResolver.ContinuousDeploy()
}

// Environments
func (r *ProjectResolver) Environments() []*EnvironmentResolver {
	db_resolvers := r.DBProjectResolver.Environments()
	gql_resolvers := make([]*EnvironmentResolver, 0, len(db_resolvers))

	for _, i := range db_resolvers {
		gql_resolvers = append(gql_resolvers, &EnvironmentResolver{DBEnvironmentResolver: i})
	}

	return gql_resolvers
}

// Bookmarked
func (r *ProjectResolver) Bookmarked(ctx context.Context) bool {
	return r.DBProjectResolver.Bookmarked(ctx)
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
