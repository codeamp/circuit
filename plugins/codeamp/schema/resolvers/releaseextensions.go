package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) ReleaseExtensions(ctx context.Context) ([]*ReleaseExtensionResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.ReleaseExtension
	var results []*ReleaseExtensionResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, releaseExtension := range rows {
		results = append(results, &ReleaseExtensionResolver{db: r.db, ReleaseExtension: releaseExtension})
	}

	return results, nil
}
