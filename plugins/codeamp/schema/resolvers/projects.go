package codeamp_schema_resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
)

func (r *Resolver) Projects(ctx context.Context) ([]*ProjectResolver, error) {
	var rows []codeamp_models.Project
	var results []*ProjectResolver

	r.DB.Find(&rows)
	for _, project := range rows {
		results = append(results, &ProjectResolver{DB: r.DB, Project: project})
	}

	return results, nil
}
