package codeamp_schema_resolvers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

func (r *Resolver) UserAuth(ctx context.Context, args *struct {
	Email    string
	Password string
}) (*UserAuthResolver, error) {
	var user codeamp_models.User

	if r.DB.Where("email = ?", args.Email).First(&user).RecordNotFound() {
		return nil, errors.New("User not found")
	}

	if !utils.CheckPasswordHash(args.Password, user.Password) {
		return nil, errors.New("Authentication failed")
	}

	r.DB.Model(user).Association("Permissions").Find(&user.Permissions)

	return &UserAuthResolver{DB: r.DB, User: user}, nil
}

type UserAuthResolver struct {
	DB   *gorm.DB
	User codeamp_models.User
}

func (r *UserAuthResolver) Token() (string, error) {
	var permissions []string

	permissions = append(permissions, fmt.Sprintf("user:%s", r.User.Model.ID))

	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	claims := utils.JWTClaims{
		UserId:      r.User.Model.ID.String(),
		Permissions: permissions,
		StandardClaims: jwt.StandardClaims{
			Issuer:    viper.GetString("plugins.codeamp.jwt_issuer"),
			IssuedAt:  time.Now().UTC().Unix(),
			ExpiresAt: time.Now().Add(time.Minute * 60).Unix(),
		},
	}

	key := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	token, err := key.SignedString([]byte(viper.GetString("plugins.codeamp.jwt_secret")))
	if err != nil {
		return "", err
	}

	return token, nil
}
