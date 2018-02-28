package resolvers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/schema/scalar"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

// UserInput
type UserInput struct {
	Email    string
	Password string
}

// UserPermissionsInput
type UserPermissionsInput struct {
	UserId      string
	Permissions []scalar.Json
}

// PermissionInput used when parsing
// permission objects from request payload
type PermissionInput struct {
	Permission string `json:"permission"`
	Checked    bool   `json:"checked"`
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
	var err error
	invalidPermissionsErr := "Input parameters do not match distinct user permission scopes."
	invalidUserIdErr := "User ID is not a valid UUID."

	// Check if UserId input is valid
	userId := uuid.FromStringOrNil(args.UserPermissions.UserId)
	if userId == uuid.Nil {
		return nil, errors.New(invalidUserIdErr)
	}

	if r.db.Where("id = ?", userId).Find(&models.User{}).RecordNotFound() {
		return nil, errors.New(invalidUserIdErr)
	}

	// Filter permission args to be a distinct array
	// and unmarshal into an array of PermissionInput objects
	distinctPermissionInputs := []PermissionInput{}
	visited := make(map[string]bool)
	for _, permission := range args.UserPermissions.Permissions {
		permissionInputStruct := PermissionInput{}
		err = json.Unmarshal(permission.RawMessage, &permissionInputStruct)
		if err != nil {
			return nil, err
		}

		// Confirm there is > 1 active row with this permission value if Checked is false
		// or else do not include for deletion/ creation
		samePermissionRows := []models.UserPermission{}
		r.db.Where("value = ?", permissionInputStruct.Permission).Find(&samePermissionRows)
		if (!permissionInputStruct.Checked && len(samePermissionRows) > 1) || permissionInputStruct.Checked {
			if !visited[permissionInputStruct.Permission] {
				visited[permissionInputStruct.Permission] = true
				distinctPermissionInputs = append(distinctPermissionInputs, permissionInputStruct)
			}
		}
	}

	// Check that all the input Permissions params
	// are valid and within the distinct set of permission values
	// Query the set of permission values
	// by getting all 'distinct' values from user_permissions table
	distinctUserPermissionSet := []models.UserPermission{}
	var filteredPermissions []string // Permissions that the request user has the authority to delete
	r.db.Select("DISTINCT(value)").Find(&distinctUserPermissionSet)
	// Loop through and see if any of the input params
	// don't match the given values. We also filter permissions by what the user
	// making the request has access to
	for _, inputPermission := range distinctPermissionInputs {
		matched := false
		for _, distinctUserPermission := range distinctUserPermissionSet {
			if inputPermission.Permission == distinctUserPermission.Value {
				if _, err = utils.CheckAuth(ctx, []string{distinctUserPermission.Value}); err == nil {
					matched = true
					filteredPermissions = append(filteredPermissions, distinctUserPermission.Value)
				} else {
					return nil, err
				}
			}
		}
		if !matched {
			return nil, fmt.Errorf(invalidPermissionsErr)
		}
	}

	// First, we delete all the user_permissions rows related to the user_id input
	// that the request user has authority to modify
	// Then, we create user_permissions rows for each of the Permissions inputs
	userPermission := models.UserPermission{}
	for _, filteredPermission := range filteredPermissions {
		// check if user has this permission
		r.db.Where("user_id = ? and value = ?", userId, filteredPermission).Find(&userPermission)
		r.db.Delete(&userPermission)
	}
	var results []string
	for _, inputPermission := range distinctPermissionInputs {
		// Create the 'checked' permissions
		if inputPermission.Checked {
			userPermissionRow := models.UserPermission{
				UserId: userId,
				Value:  inputPermission.Permission,
			}
			r.db.Create(&userPermissionRow)
			results = append(results, userPermissionRow.Value)
		}
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

	// permissions = append(permissions, fmt.Sprintf("user:%s", r.User.Model.ID))
	for _, permission := range r.User.Permissions {
		permissions = append(permissions, permission.Value)
	}

	return permissions
}

func (r *UserResolver) Created() graphql.Time {
	return graphql.Time{Time: r.User.Model.CreatedAt}
}
