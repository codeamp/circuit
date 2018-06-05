package db_resolver

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/transistor"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

type UserResolver struct {
	model.User
	DB *gorm.DB
}

// Queries
// func (u *User) User() {

// }

// ID
func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.User.Model.ID.String())
}

// Email
func (r *UserResolver) Email() string {
	return r.User.Email
}

// Permissions
func (r *UserResolver) Permissions(ctx context.Context) []string {
	if _, err := CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil
	}

	var permissions []string

	r.DB.Model(r.User).Association("Permissions").Find(&r.User.Permissions)

	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}

// Created
func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.User.Model.CreatedAt}
}

func (r *UserResolver) MarshalJSON() ([]byte, error) {
	return json.Marshal(&r.User)
}

func (r *UserResolver) UnmarshalJSON(data []byte) error {
	return json.Unmarshal(data, &r.User)
}

//Claims
type Claims struct {
	UserID      string   `json:"userID"`
	Email       string   `json:"email"`
	Verified    bool     `json:"email_verified"`
	Groups      []string `json:"groups"`
	Permissions []string `json:"permissions"`
	TokenError  string   `json:"tokenError"`
}

func CheckAuth(ctx context.Context, scopes []string) (string, error) {
	claims := ctx.Value("jwt").(Claims)

	if claims.UserID == "" {
		return "", errors.New(claims.TokenError)
	}

	if transistor.SliceContains("admin", claims.Permissions) {
		return claims.UserID, nil
	}

	if len(scopes) == 0 {
		return claims.UserID, nil
	} else {
		for _, scope := range scopes {
			level := 0
			levels := strings.Count(scope, "/")

			if levels > 0 {
				for level < levels {
					if transistor.SliceContains(scope, claims.Permissions) {
						return claims.UserID, nil
					}
					scope = scope[0:strings.LastIndexByte(scope, '/')]
					level += 1
				}
			} else {
				if transistor.SliceContains(scope, claims.Permissions) {
					return claims.UserID, nil
				}
			}
		}
		return claims.UserID, errors.New("you dont have permission to access this resource")
	}

	return claims.UserID, nil
}
