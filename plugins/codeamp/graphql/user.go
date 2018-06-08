package graphql_resolver

import (
	"context"
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// User resolver
type UserResolver struct {
	DBUserResolver *db_resolver.UserResolver
}

// ID
func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.DBUserResolver.UserModel.Model.ID.String())
}

// Email
func (r *UserResolver) Email() string {
	return r.DBUserResolver.UserModel.Email
}

// Permissions
func (r *UserResolver) Permissions(ctx context.Context) []string {
	return r.DBUserResolver.Permissions(ctx)
}

// Created
func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.DBUserResolver.UserModel.Model.CreatedAt}
}

func (r *UserResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.DBUserResolver.UserModel)
}

func (r *UserResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.DBUserResolver.UserModel)
}
