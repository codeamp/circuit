package db_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
)

type ProjectExtensionResolver struct {
	model.ProjectExtension
	DB *gorm.DB
}

// Project
func (r *ProjectExtensionResolver) Project() *ProjectResolver {
	var project model.Project

	r.DB.Model(r.ProjectExtension).Related(&project)

	return &ProjectResolver{DB: r.DB, Project: project}
}

// Extension
func (r *ProjectExtensionResolver) Extension() *ExtensionResolver {
	var ext model.Extension

	r.DB.Model(r.ProjectExtension).Related(&ext)

	return &ExtensionResolver{DB: r.DB, Extension: ext}
}

// Environment
func (r *ProjectExtensionResolver) Environment() (*EnvironmentResolver, error) {
	var environment model.Environment
	if r.DB.Where("id = ?", r.ProjectExtension.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}
