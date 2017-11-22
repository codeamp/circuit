package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) Extensions(ctx context.Context) ([]*ExtensionResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.Extension
	var results []*ExtensionResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, extension := range rows {
		results = append(results, &ExtensionResolver{db: r.db, Extension: extension})
	}

	return results, nil
}
