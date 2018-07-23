package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type EnvironmentResolver struct {
	model.Project
	model.Environment
	DB *gorm.DB
}

func (r *EnvironmentResolver) Projects() []*ProjectResolver {
	var permissions []model.ProjectEnvironment
	var results []*ProjectResolver

	r.DB.Debug().Where("environment_id = ?", r.Environment.ID).Find(&permissions)
	for _, permission := range permissions {
		var project model.Project
		if !r.DB.Where("id = ?", permission.ProjectID).First(&project).RecordNotFound() {
			results = append(results, &ProjectResolver{DB: r.DB, Project: project, Environment: r.Environment})
		}
	}

	return results
}
