package graphql_resolver

import (
	"encoding/json"
	"fmt"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	graphql "github.com/graph-gophers/graphql-go"
)

func GetSecretScope(s string) model.SecretScope {
	secretScopes := []string{
		"project",
		"extension",
		"global",
	}

	for _, secretScope := range secretScopes {
		if s == secretScope {
			return model.SecretScope(secretScope)
		}
	}

	log.Info(fmt.Sprintf("SecretScope not found: %s", s))

	return model.SecretScope("unknown")
}

// SecretResolver resolver for Secret
type SecretResolver struct {
	DBSecretResolver *db_resolver.SecretResolver
}

// ID
func (r *SecretResolver) ID() graphql.ID {
	return graphql.ID(r.DBSecretResolver.Secret.Model.ID.String())
}

// Key
func (r *SecretResolver) Key() string {
	return r.DBSecretResolver.Secret.Key
}

// Value
func (r *SecretResolver) Value() string {
	return r.DBSecretResolver.Value()
}

// Scope
func (r *SecretResolver) Scope() string {
	return string(r.DBSecretResolver.Secret.Scope)
}

// Project
func (r *SecretResolver) Project() *ProjectResolver {
	return &ProjectResolver{DBProjectResolver: r.DBSecretResolver.Project()}
}

// User
func (r *SecretResolver) User() *UserResolver {
	return &UserResolver{DBUserResolver: r.DBSecretResolver.User()}
}

// Type
func (r *SecretResolver) Type() string {
	return string(r.DBSecretResolver.Secret.Type)
}

// Versions
func (r *SecretResolver) Versions() ([]*SecretResolver, error) {
	db_resolvers, err := r.DBSecretResolver.Versions()
	gql_resolvers := make([]*SecretResolver, 0, len(db_resolvers))

	for _, i := range db_resolvers {
		gql_resolvers = append(gql_resolvers, &SecretResolver{DBSecretResolver: i})
	}

	return gql_resolvers, err
}

// Environment
func (r *SecretResolver) Environment() *EnvironmentResolver {
	return &EnvironmentResolver{DBEnvironmentResolver: r.DBSecretResolver.Environment()}
}

// Created
func (r *SecretResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBSecretResolver.Secret.Model.CreatedAt}
}

// IsSecret
func (r *SecretResolver) IsSecret() bool {
	return r.DBSecretResolver.IsSecret()
}

func (r *SecretResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBSecretResolver.Secret)
}

func (r *SecretResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBSecretResolver.Secret)
}
