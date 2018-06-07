package graphql_resolver

import (
	"context"
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/db"
	graphql "github.com/graph-gophers/graphql-go"
)

// User resolver
type UserResolver struct {
	*db_resolver.UserResolver
}

// ID
func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.UserResolver.User.Model.ID.String())
}

// Email
func (r *UserResolver) Email() string {
	return r.UserResolver.User.Email
}

// Permissions
func (r *UserResolver) Permissions(ctx context.Context) []string {
	if _, err := db_resolver.CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil
	}

	var permissions []string

	r.DB.Model(r.User).Association("Permissions").Find(&r.UserResolver.User.Permissions)

	for _, permission := range r.UserResolver.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}

// Created
func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.UserResolver.User.Model.CreatedAt}
}

func (r *UserResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.UserResolver.User)
}

func (r *UserResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.UserResolver.User)
}

func (r *UserResolver) User() {

}

func (r *UserResolver) Error() error {
	return nil
}
