package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) ExtensionSpecs(ctx context.Context) ([]*ExtensionSpecResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.ExtensionSpec
	var results []*ExtensionSpecResolver

	r.db.Order("created desc").Find(&rows)
	for _, extensionSpec := range rows {
		results = append(results, &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec})
	}

	return results, nil
}
