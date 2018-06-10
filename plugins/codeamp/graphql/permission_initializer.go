package graphql_resolver

import (
	"context"
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/auth"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Extension Resolver Initializer
type PermissionResolverInitializer struct {
	DB *gorm.DB
}

// Permissions
func (r *PermissionResolverInitializer) Permissions(ctx context.Context) (model.JSON, error) {
	var rows []model.UserPermission
	var results = make(map[string]bool)

	r.DB.Unscoped().Select("DISTINCT(value)").Find(&rows)

	for _, userPermission := range rows {
		if _, err := auth.CheckAuth(ctx, []string{userPermission.Value}); err != nil {
			results[userPermission.Value] = false
		} else {
			results[userPermission.Value] = true
		}
	}

	bytes, err := json.Marshal(results)
	if err != nil {
		return model.JSON{}, err
	}

	return model.JSON{bytes}, nil
}
