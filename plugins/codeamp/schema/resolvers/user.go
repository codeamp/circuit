package codeamp_schema_resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type UserInput struct {
	Email    string
	Password string
}

func (r *Resolver) CreateUser(args *struct{ User *UserInput }) *UserResolver {
	passwordHash, _ := utils.HashPassword(args.User.Password)

	user := codeamp_models.User{
		Email:    args.User.Email,
		Password: passwordHash,
	}

	r.DB.Create(&user)

	return &UserResolver{DB: r.DB, User: user}
}

func (r *Resolver) User(ctx context.Context, args *struct{ ID graphql.ID }) (*UserResolver, error) {
	var user codeamp_models.User

	if err := utils.CheckAuth(ctx, []string{"admin", fmt.Sprintf("user:%s", args.ID)}); err != nil {
		return nil, err
	}

	if err := r.DB.Where("id = ?", args.ID).First(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{DB: r.DB, User: user}, nil
}

type UserResolver struct {
	DB   *gorm.DB
	User codeamp_models.User
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.User.Model.ID.String())
}

func (r *UserResolver) Email(ctx context.Context) (string, error) {
	return r.User.Email, nil
}

func (r *UserResolver) Permissions() []string {
	var permissions []string

	r.DB.Model(r.User).Association("Permissions").Find(&r.User.Permissions)

	permissions = append(permissions, fmt.Sprintf("user:%s", r.User.Model.ID))

	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}
