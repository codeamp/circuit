package graphql_resolver

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	graphql "github.com/graph-gophers/graphql-go"
	"github.com/jinzhu/gorm"
)

// User Resolver Query
type UserResolverQuery struct {
	DB *gorm.DB
}

func (u *UserResolverQuery) User(ctx context.Context, args *struct {
	ID *graphql.ID
}) (*UserResolver, error) {
	var userID string

	if args.ID != nil {
		userID = string(*args.ID)
	} else {
		claims := ctx.Value("jwt").(model.Claims)
		userID = claims.UserID
	}

	var err error
	if _, err = auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
		return nil, err
	}

	resolver := UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB}}
	if err = u.DB.Where("id = ?", userID).First(&resolver.DBUserResolver.User).Error; err != nil {
		return nil, err
	}

	return &resolver, nil
}

func (u *UserResolverQuery) Users(ctx context.Context) ([]*UserResolver, error) {
	var rows []model.User
	var results []*UserResolver

	u.DB.Order("created_at desc").Find(&rows)

	for _, user := range rows {
		results = append(results, &UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB, User: user}})
	}

	return results, nil
}

// Permissions
func (r *UserResolverQuery) Permissions(ctx context.Context) (model.JSON, error) {
	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return model.JSON{}, err
	}

	var rows []model.UserPermission
	var results = make(map[string]bool)

	r.DB.Unscoped().Select("DISTINCT(value)").Find(&rows)

	for _, userPermission := range rows {
		if _, err := auth.CheckAuth(ctx, []string{userPermission.Value}); err != nil {
			results[userPermission.Value] = false
		} else {
			results[userPermission.Value] = true
		}
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		return model.JSON{}, err
	}

	return model.JSON{bytes}, nil
}
