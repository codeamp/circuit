package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) EnvironmentProjectBranches(ctx context.Context) ([]*EnvironmentResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.EnvironmentProjectBranch
	var results []*EnvironmentProjectBranchResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, envBranch := range rows {
		results = append(results, &EnvironmentProjectBranchResolver{db: r.db, EnvironmentProjectBranch: envBranch})
	}

	return results, nil
}
