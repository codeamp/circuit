package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) Environments(ctx context.Context) ([]*EnvironmentResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.Environment
	var results []*EnvironmentResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, env := range rows {
		results = append(results, &EnvironmentResolver{db: r.db, Environment: env})
	}

	return results, nil
}
