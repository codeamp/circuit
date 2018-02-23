package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type UserInput struct {
	Email    string
	Password string
}

type UserPermissionsInput struct {
	UserId      string
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

func (r *Resolver) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissions *UserPermissionsInput }) ([]string, error) {
	invalidPermissionsErr := "Input parameters do not match distinct user permission scopes."
	invalidUserIdErr := "User ID is not a valid UUID."

	// Check if UserId input is valid
	userId := uuid.FromStringOrNil(args.UserPermissions.UserId)
	if userId == uuid.Nil {
		return nil, fmt.Errorf(invalidUserIdErr)
	}

	if r.db.Where("id = ?").Find(&models.User{}).RecordNotFound() {
		return nil, fmt.Errorf(invalidUserIdErr)
	}

	// Check that all the input Permissions params
	// are valid and within the distinct set of permission values
	// Query the set of permission values
	// by getting all 'distinct' values from user_permissions table
	var distinctUserPermissionSet []models.UserPermission
	var filteredPermissions []models.UserPermission // Permissions that the request user has the authority to delete
	r.db.Select("DISTINCT(value)").Find(&distinctUserPermissionSet)

	// Loop through and see if any of the input params
	// don't match the given values. We also filter permissions by what the user
	// making the request has access to
	var found bool
	for _, inputPermission := range args.UserPermissions.Permissions {
		found = false
		for _, distinctUserPermission := range distinctUserPermissionSet {
			_, err := utils.CheckAuth(ctx, []string{"admin", distinctUserPermission.Value})
			if inputPermission == distinctUserPermission.Value && err == nil {
				found = true
				filteredPermissions = append(filteredPermissions, distinctUserPermission)
			}
		}

		if !found {
			return nil, fmt.Errorf(invalidPermissionsErr)
		}

	}

	// First, we delete all the user_permissions rows related to the user_id input
	// that the request user has authority to modify
	// Then, we create user_permissions rows for each of the Permissions inputs
	r.db.Delete(&filteredPermissions)
	var results []string
	for _, inputPermission := range args.UserPermissions.Permissions {
		userPermissionRow := models.UserPermission{
			UserId: userId,
			Value:  inputPermission,
		}
		r.db.Create(&userPermissionRow)
		results = append(results, userPermissionRow.Value)
	}

	return results, nil
}

func (r *Resolver) User(ctx context.Context, args *struct{ ID *graphql.ID }) (*UserResolver, error) {
	var err error
	var userId string
	var user models.User
	if userId, err = utils.CheckAuth(ctx, []string{"admin", fmt.Sprintf("user:%s", args.ID)}); err != nil {
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
