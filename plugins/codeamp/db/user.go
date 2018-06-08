package db_resolver

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type UserResolver struct {
	UserModel model.User
	DB        *gorm.DB
}

func (u *UserResolver) Permissions(ctx context.Context) []string {
	if _, err := auth.CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil
	}

	var permissions []string

	u.DB.Model(u.UserModel).Association("Permissions").Find(&u.UserModel.Permissions)

	for _, permission := range u.UserModel.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}
