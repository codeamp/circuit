package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type FeatureResolver struct {
	model.Feature
	DB *gorm.DB
}

func (r *FeatureResolver) Project() *ProjectResolver {
	var project model.Project
	r.DB.Model(r.Feature).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}
