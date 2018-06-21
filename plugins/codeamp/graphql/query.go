package graphql_resolver

import (
	"context"
	_ "encoding/json"

	graphql "github.com/graph-gophers/graphql-go"

	_ "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	_ "github.com/codeamp/logger"
)

// User
func (r *Resolver) User(ctx context.Context, args *struct {
	ID *graphql.ID
}) (*UserResolver, error) {
	initializer := UserResolverQuery{DB: r.DB}
	return initializer.User(ctx, args)
}

// Users
func (r *Resolver) Users(ctx context.Context) ([]*UserResolver, error) {
	initializer := UserResolverQuery{DB: r.DB}
	return initializer.Users(ctx)
}

// Project
func (r *Resolver) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentID *string
}) (*ProjectResolver, error) {
	initializer := ProjectResolverQuery{DB: r.DB}
	return initializer.Project(ctx, args)
}

// Projects
func (r *Resolver) Projects(ctx context.Context, args *struct {
	ProjectSearch *model.ProjectSearchInput
}) ([]*ProjectResolver, error) {
	initializer := ProjectResolverQuery{DB: r.DB}
	return initializer.Projects(ctx, args.ProjectSearch)
}

func (r *Resolver) Features(ctx context.Context) ([]*FeatureResolver, error) {
	initializer := FeatureResolverQuery{DB: r.DB}
	return initializer.Features(ctx)
}

func (r *Resolver) Services(ctx context.Context) ([]*ServiceResolver, error) {
	initializer := ServiceResolverQuery{DB: r.DB}
	return initializer.Services(ctx)
}

func (r *Resolver) ServiceSpecs(ctx context.Context) ([]*ServiceSpecResolver, error) {
	initializer := ServiceSpecResolverQuery{DB: r.DB}
	return initializer.ServiceSpecs(ctx)
}

func (r *Resolver) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	initializer := ReleaseResolverQuery{DB: r.DB}
	return initializer.Releases(ctx)
}

func (r *Resolver) Environments(ctx context.Context, args *struct{ ProjectSlug *string }) ([]*EnvironmentResolver, error) {
	initializer := EnvironmentResolverQuery{DB: r.DB}
	return initializer.Environments(ctx, args)
}

func (r *Resolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	initializer := SecretResolverQuery{DB: r.DB}
	return initializer.Secrets(ctx)
}

func (r *Resolver) Extensions(ctx context.Context, args *struct{ EnvironmentID *string }) ([]*ExtensionResolver, error) {
	initializer := ExtensionResolverQuery{DB: r.DB}
	return initializer.Extensions(ctx, args)
}

func (r *Resolver) ProjectExtensions(ctx context.Context) ([]*ProjectExtensionResolver, error) {
	initializer := ProjectExtensionResolverQuery{DB: r.DB}
	return initializer.ProjectExtensions(ctx)
}

func (r *Resolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	initializer := ReleaseExtensionResolverQuery{DB: r.DB}
	return initializer.ReleaseExtensions(ctx)
}

// Permissions
func (r *Resolver) Permissions(ctx context.Context) (model.JSON, error) {
	initializer := PermissionResolverQuery{DB: r.DB}
	return initializer.Permissions(ctx)
}
