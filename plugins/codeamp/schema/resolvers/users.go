package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) Users(ctx context.Context) ([]*UserResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{"admin"}); err != nil {
		return nil, err
	}

	var rows []models.User
	var results []*UserResolver

	r.db.Order("created_at desc").Find(&rows)

	for _, user := range rows {
		results = append(results, &UserResolver{db: r.db, User: user})
	}

	return results, nil
}

func (r *Resolver) Permissions(ctx context.Context) ([]string, error) {
	var rows []models.UserPermission
	var results []string

	r.db.Select("DISTINCT(value)").Find(&rows)

	for _, distinctUserPermission := range rows {
		// Check if user doing the query has access to view the permission
		if _, err := utils.CheckAuth(ctx, []string{distinctUserPermission.Value}); err != nil {
			return nil, err
		}

		results = append(results, distinctUserPermission.Value)
	}

	return results, nil
}
