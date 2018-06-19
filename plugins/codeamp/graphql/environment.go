package graphql_resolver

import (
	"encoding/json"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// EnvironmentResolver resolver for Environment
type EnvironmentResolver struct {
	DBEnvironmentResolver *db_resolver.EnvironmentResolver
}

// ID
func (r *EnvironmentResolver) ID() graphql.ID {
	return graphql.ID(r.DBEnvironmentResolver.Environment.Model.ID.String())
}

// Name
func (r *EnvironmentResolver) Name() string {
	return r.DBEnvironmentResolver.Environment.Name
}

// Color
func (r *EnvironmentResolver) Color() string {
	return r.DBEnvironmentResolver.Environment.Color
}

// Key
func (r *EnvironmentResolver) Key() string {
	return r.DBEnvironmentResolver.Environment.Key
}

// Is Default
func (r *EnvironmentResolver) IsDefault() bool {
	return r.DBEnvironmentResolver.Environment.IsDefault
}

// Projects - get projects permissioned for the environment
func (r *EnvironmentResolver) Projects() []*ProjectResolver {
	db_resolvers := r.DBEnvironmentResolver.Projects()
	gql_resolvers := make([]*ProjectResolver, 0, len(db_resolvers))

	for _, i := range db_resolvers {
		gql_resolvers = append(gql_resolvers, &ProjectResolver{DBProjectResolver: i})
	}

	return gql_resolvers
}

// Created
func (r *EnvironmentResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBEnvironmentResolver.Environment.Model.CreatedAt}
}

func (r *EnvironmentResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBEnvironmentResolver.Environment)
}

func (r *EnvironmentResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBEnvironmentResolver.Environment)
}
