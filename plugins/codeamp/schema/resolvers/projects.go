package codeamp_schema_resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	"github.com/codeamp/circuit/plugins/codeamp/utils"
)

func (r *Resolver) Projects(ctx context.Context) ([]*ProjectResolver, error) {
	if _, err := utils.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	var rows []codeamp_models.Project
	var results []*ProjectResolver

	r.DB.Find(&rows)
	for _, project := range rows {
		results = append(results, &ProjectResolver{DB: r.DB, Project: project})
	}

	return results, nil
}
