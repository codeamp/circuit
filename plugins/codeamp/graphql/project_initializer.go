package graphql_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// User Resolver Initializer
type ProjectResolverInitializer struct {
	Proj model.Project
	DB   *gorm.DB
}

func (u *ProjectResolverInitializer) Project(ctx context.Context, userID string) (*ProjectResolver, error) {
	// var err error
	// if _, err = auth.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
	// 	return nil, err
	// }

	// resolver := UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB}}
	// if err = u.DB.Where("id = ?", userID).First(&resolver.DBUserResolver.UserModel).Error; err != nil {
	// 	return nil, err
	// }

	// return &resolver, nil
	return nil, nil
}

func (u *ProjectResolverInitializer) Projects(ctx context.Context) ([]*ProjectResolver, error) {
	// var rows []model.User
	// var results []*UserResolver

	// u.DB.Order("created_at desc").Find(&rows)

	// for _, user := range rows {
	// 	results = append(results, &UserResolver{DBUserResolver: &db_resolver.UserResolver{DB: u.DB, UserModel: user}})
	// }

	// return results, nil
	return nil, nil
}
