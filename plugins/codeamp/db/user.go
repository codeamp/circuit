package db_resolver

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type UserResolver struct {
	model.User
	DB *gorm.DB
}

// Queries
func (u *UserResolver) QueryUser(ctx context.Context, userID string) error {
	var err error
	if _, err = CheckAuth(ctx, []string{fmt.Sprintf("user/%s", userID)}); err != nil {
		return err
	}

	if err = u.DB.Where("id = ?", userID).First(&u.User).Error; err != nil {
		return err
	}

	return nil
}

func (u *UserResolver) QueryUsers() {

}
