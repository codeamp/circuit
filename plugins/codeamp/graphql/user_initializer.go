package graphql_resolver

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// User Resolver Initializer
type UserResolverInitializer struct {
	DB *gorm.DB
}

func (u *UserResolverInitializer) User(ctx context.Context, userID string) (*UserResolver, error) {
	var err error
	if _, err = auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
		return nil, err
	}

	resolver := UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB}}
	if err = u.DB.Where("id = ?", userID).First(&resolver.DBUserResolver.UserModel).Error; err != nil {
		return nil, err
	}

	return &resolver, nil
}

func (u *UserResolverInitializer) Users(ctx context.Context) ([]*UserResolver, error) {
	var rows []model.User
	var results []*UserResolver

	u.DB.Order("created_at desc").Find(&rows)

	for _, user := range rows {
		results = append(results, &UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB, UserModel: user}})
	}

	return results, nil
}
