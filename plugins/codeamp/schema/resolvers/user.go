package resolvers

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

type UserPermissionsInput struct {
	UserId string
	Permissions []string
}

func (r *Resolver) CreateUser(args *struct{ User *UserInput }) *UserResolver {
	passwordHash, _ := utils.HashPassword(args.User.Password)

	user := models.User{
		Email:    args.User.Email,
		Password: passwordHash,
	}

	r.db.Create(&user)

	return &UserResolver{db: r.db, User: user}
}

func (r *UserResolver) UpdateUserPermissions(args *struct{ UserPermissions *UserPermissionsInput}) (string, error) {
	// Check that all the input params
	// are valid and within the distinct set of permission values

	// Query the set of permission values
	// by getting all 'distinct' values from user_permissions table
	var distinctUserPermissionSet []models.UserPermission{}
	errorMessage := "Input parameters do not match distinct user permission scopes."
	if r.db.Where("distinct").Find(&distinctUserPermissionSet).RecordNotFound() {
		log.Info(errorMessage)
		return nil, fmt.Errorf(errorMessage)
	}

	// Loop through and see if any of the input params
	// don't match the given values
	var found bool
	for _, inputPermission := range args.UserPermissions.Permissions {
		found = false
		for _, distinctUserPermission := range distinctUserPermissionSet {
			if inputPermission == distinctUserPermission {
				found = true
			}
		}
		if !found {
			return nil, fmt.Errorf(errorMessage)
		}
	}

	// First, we delete all the user_permissions rows related to the user_id input
	// Then, we create user_permissions rows for each of the Permissions inputs
	var resp []models.UserPermission{}
	r.db.Delete("user_id = ?", args.UserPermissions.UserId)
	for _, inputPermission := range args.UserPermissions.Permissions {
		userPermissionRow := models.UserPermission{
			UserId: args.UserPermissions.UserId,
			Value: inputPermission
		}
		r.db.Create(&userPermissionRow)
		append(resp, userPermissionRow)
	}

	return resp, nil
}

func (r *Resolver) User(ctx context.Context, args *struct{ ID *graphql.ID }) (*UserResolver, error) {
	var err error
	var userId string
	var user models.User
	if userId, err = utils.CheckAuth(ctx, []string{fmt.Sprintf("user/%s", args.ID)}); err != nil {
		return nil, err
	}

	if args.ID != nil {
		userId = string(*args.ID)
	}

	if err := r.db.Where("id = ?", userId).First(&user).Error; err != nil {
		return nil, err
	}

	return &UserResolver{db: r.db, User: user}, nil
}

type UserResolver struct {
	db   *gorm.DB
	User models.User
}

func (r *UserResolver) ID() graphql.ID {
	return graphql.ID(r.User.Model.ID.String())
}

func (r *UserResolver) Email(ctx context.Context) (string, error) {
	return r.User.Email, nil
}

func (r *UserResolver) Permissions() []string {
	var permissions []string

	r.db.Model(r.User).Association("Permissions").Find(&r.User.Permissions)

	permissions = append(permissions, fmt.Sprintf("user:%s", r.User.Model.ID))

	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}

func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.User.Model.CreatedAt}
}