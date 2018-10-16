package graphql_resolver

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

// User Resolver Mutation
type UserResolverMutation struct {
	DB *gorm.DB
}

func (r *UserResolverMutation) UpdateUserPermissions(ctx context.Context, args *struct{ UserPermissions *model.UserPermissionsInput }) ([]string, error) {
	var err error
	var results []string

	if r.DB.Where("id = ?", args.UserPermissions.UserID).Find(&model.User{}).RecordNotFound() {
		return nil, fmt.Errorf("User not found")
	}

	for _, permission := range args.UserPermissions.Permissions {
		if _, err = auth.CheckAuth(ctx, []string{permission.Value}); err != nil {
			return nil, err
		}
	}

	for _, permission := range args.UserPermissions.Permissions {
		if permission.Grant == true {
			userPermission := model.UserPermission{
				UserID: uuid.FromStringOrNil(args.UserPermissions.UserID),
				Value:  permission.Value,
			}
			r.DB.Where(userPermission).FirstOrCreate(&userPermission)
			results = append(results, permission.Value)
		} else {
			r.DB.Where("user_id = ? AND value = ?", args.UserPermissions.UserID, permission.Value).Delete(&model.UserPermission{})
		}
	}

	return results, nil
}
