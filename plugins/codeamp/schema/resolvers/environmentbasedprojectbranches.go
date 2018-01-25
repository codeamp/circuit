package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) EnvironmentBasedProjectBranches(ctx context.Context) ([]*EnvironmentBasedProjectBranchResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.EnvironmentBasedProjectBranch
	var results []*EnvironmentBasedProjectBranchResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, envBranch := range rows {
		results = append(results, &EnvironmentBasedProjectBranchResolver{db: r.db, EnvironmentBasedProjectBranch: envBranch})
	}

	return results, nil
}
