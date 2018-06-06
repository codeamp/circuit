package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type UserResolver struct {
	model.User
	DB *gorm.DB
}

// Queries
func (u *UserResolver) QueryUser() {

}

func (u *UserResolver) QueryUsers() {

}
