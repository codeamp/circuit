package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) ExtensionSpecEnvironmentVariables(ctx context.Context) ([]*ExtensionSpecEnvironmentVariableResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.ExtensionSpecEnvironmentVariable
	var results []*ExtensionSpecEnvironmentVariableResolver

	r.db.Order("created_at desc").Find(&rows)
	for _, extensionSpecEnvironmentVariable := range rows {
		results = append(results, &ExtensionSpecEnvironmentVariableResolver{db: r.db, ExtensionSpecEnvironmentVariable: extensionSpecEnvironmentVariable})
	}

	return results, nil
}
