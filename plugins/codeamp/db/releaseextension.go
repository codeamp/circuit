package db_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
)

type ReleaseExtensionResolver struct {
	model.ReleaseExtension
	DB *gorm.DB
}

// Release
func (r *ReleaseExtensionResolver) Release() (*ReleaseResolver, error) {
	release := model.Release{}

	if r.DB.Where("id = ?", r.ReleaseExtension.ReleaseID.String()).Find(&release).RecordNotFound() {
		log.InfoWithFields("extension not found", log.Fields{
			"id": r.ReleaseExtension.ReleaseID.String(),
		})
		return &ReleaseResolver{DB: r.DB, Release: release}, fmt.Errorf("Couldn't find release")
	}

	return &ReleaseResolver{DB: r.DB, Release: release}, nil
}

// ProjectExtension
func (r *ReleaseExtensionResolver) Extension() (*ProjectExtensionResolver, error) {
	extension := model.ProjectExtension{}

	if r.DB.Unscoped().Where("id = ?", r.ReleaseExtension.ProjectExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("extension not found", log.Fields{
			"id": r.ReleaseExtension.ProjectExtensionID,
		})
		return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, fmt.Errorf("Couldn't find extension")
	}

	return &ProjectExtensionResolver{DB: r.DB, ProjectExtension: extension}, nil
}
