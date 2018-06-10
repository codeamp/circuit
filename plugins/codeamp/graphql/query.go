package graphql_resolver

import (
	"context"
	_ "encoding/json"
	"fmt"

	graphql "github.com/graph-gophers/graphql-go"
	uuid "github.com/satori/go.uuid"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	_ "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	_ "github.com/codeamp/logger"
)

// User
func (r *Resolver) User(ctx context.Context, args *struct {
	ID *graphql.ID
}) (*UserResolver, error) {
	var userID string

	if args.ID != nil {
		userID = string(*args.ID)
	} else {
		claims := ctx.Value("jwt").(model.Claims)
		userID = claims.UserID
	}

	initializer := UserResolverInitializer{DB: r.DB}
	resolver, err := initializer.User(ctx, userID)

	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// Users
func (r *Resolver) Users(ctx context.Context) ([]*UserResolver, error) {
	initializer := UserResolverInitializer{DB: r.DB}
	resolvers, err := initializer.Users(ctx)

	return resolvers, err
}

// Project
func (r *Resolver) Project(ctx context.Context, args *struct {
	ID            *graphql.ID
	Slug          *string
	Name          *string
	EnvironmentID *string
}) (*ProjectResolver, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	identifier := make(map[string]string)
	if args.ID != nil {
		identifier["ID"] = string(*args.ID)
	} else if args.Slug != nil {
		identifier["Slug"] = *args.Slug
	} else if args.Name != nil {
		identifier["Name"] = *args.Name
	} else {
		return nil, fmt.Errorf("Missing argument id or slug")
	}

	if args.EnvironmentID == nil {
		return nil, fmt.Errorf("Missing environment id")
	}

	environmentID, err := uuid.FromString(*args.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Environment ID should be of type uuid")
	}

	initializer := ProjectResolverInitializer{DB: r.DB}
	resolver, err := initializer.Project(ctx, identifier, environmentID)

	if err != nil {
		return nil, err
	}

	return resolver, nil
}

// Projects
func (r *Resolver) Projects(ctx context.Context, args *struct {
	ProjectSearch *model.ProjectSearchInput
}) ([]*ProjectResolver, error) {
	initializer := ProjectResolverInitializer{DB: r.DB}
	return initializer.Projects(ctx, args.ProjectSearch)
}

func (r *Resolver) Features(ctx context.Context) ([]*FeatureResolver, error) {
	initializer := FeatureResolverInitializer{DB: r.DB}
	return initializer.Features(ctx)
}

func (r *Resolver) Services(ctx context.Context) ([]*ServiceResolver, error) {
	initializer := ServiceResolverInitializer{DB: r.DB}
	return initializer.Services(ctx)
}

func (r *Resolver) ServiceSpecs(ctx context.Context) ([]*ServiceSpecResolver, error) {
	initializer := ServiceSpecResolverInitializer{DB: r.DB}
	return initializer.ServiceSpecs(ctx)
}

func (r *Resolver) Releases(ctx context.Context) ([]*ReleaseResolver, error) {
	initializer := ReleaseResolverInitializer{DB: r.DB}
	return initializer.Releases(ctx)
}

func (r *Resolver) Environments(ctx context.Context, args *struct{ ProjectSlug *string }) ([]*EnvironmentResolver, error) {
	initializer := EnvironmentResolverInitializer{DB: r.DB}
	return initializer.Environments(ctx, args)
}

func (r *Resolver) Secrets(ctx context.Context) ([]*SecretResolver, error) {
	initializer := SecretResolverInitializer{DB: r.DB}
	return initializer.Secrets(ctx)
}

func (r *Resolver) Extensions(ctx context.Context, args *struct{ EnvironmentID *string }) ([]*ExtensionResolver, error) {
	initializer := ExtensionResolverInitializer{DB: r.DB}
	return initializer.Extensions(ctx, args)
}

func (r *Resolver) ProjectExtensions(ctx context.Context) ([]*ProjectExtensionResolver, error) {
	initializer := ProjectExtensionResolverInitializer{DB: r.DB}
	return initializer.ProjectExtensions(ctx)
}

func (r *Resolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	initializer := ReleaseExtensionResolverInitializer{DB: r.DB}
	return initializer.ReleaseExtensions(ctx)
}

// Permissions
func (r *Resolver) Permissions(ctx context.Context) (model.JSON, error) {
	initializer := PermissionResolverInitializer{DB: r.DB}
	return initializer.Permissions(ctx)
}
