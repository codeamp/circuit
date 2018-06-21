package db_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
)

type ExtensionResolver struct {
	model.Extension
	DB *gorm.DB
}

func (r *ExtensionResolver) Environment() (*EnvironmentResolver, error) {
	environment := model.Environment{}

	if r.DB.Where("id = ?", r.Extension.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.Extension.EnvironmentID,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{DB: r.DB, Environment: environment}, nil
}
