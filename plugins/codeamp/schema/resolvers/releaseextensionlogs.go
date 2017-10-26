package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) ReleaseExtensionsLogs(ctx context.Context) ([]*ReleaseExtensionLogResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.ReleaseExtensionLog
	var results []*ReleaseExtensionLogResolver

	r.db.Order("created desc").Find(&rows)
	for _, releaseExtensionLog := range rows {
		results = append(results, &ReleaseExtensionLogResolver{db: r.db, ReleaseExtensionLog: releaseExtensionLog})
	}

	return results, nil
}
