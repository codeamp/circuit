package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) EnvironmentVariables(ctx context.Context) ([]*EnvironmentVariableResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []models.EnvironmentVariable
	var results []*EnvironmentVariableResolver

	r.db.Where("scope != ?", "project").Order("created_at desc").Find(&rows)
	for _, envVar := range rows {
		var envVarValue models.EnvironmentVariableValue
		r.db.Where("environment_variable_id = ?", envVar.Model.ID).Order("created_at desc").First(&envVarValue)
		results = append(results, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar, EnvironmentVariableValue: envVarValue})
	}

	return results, nil
}
