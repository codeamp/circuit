package db_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
)

type ServiceSpecResolver struct {
	model.ServiceSpec
	DB *gorm.DB
}

// Name
func (r *ServiceSpecResolver) Name() string {
	env := model.Environment{}
	project := model.Project{}
	service := model.Service{}

	if err := r.DB.Where("id = ?", r.ServiceSpec.ServiceID).First(&service).Error; err != nil {
		log.InfoWithFields("service not found", log.Fields{
			"id": r.ServiceSpec.ServiceID,
		})
		return r.ServiceSpec.Name
	}

	if err := r.DB.Where("id = ?", service.ProjectID).First(&project).Error; err != nil {
		log.InfoWithFields("project not found", log.Fields{
			"id": service.ProjectID,
		})
		return r.ServiceSpec.Name
	}

	if err := r.DB.Where("id = ?", service.EnvironmentID).First(&env).Error; err != nil {
		log.InfoWithFields("environment not found", log.Fields{
			"id": service.EnvironmentID,
		})
		return r.ServiceSpec.Name
	}

	return fmt.Sprintf("%s/%s/%s", service.Name, project.Slug, env.Key)
}

// Service
func (r *ServiceSpecResolver) Service() (*ServiceResolver, error) {
	service := model.Service{}
	if err := r.DB.Where("id = ?", r.ServiceSpec.ServiceID).First(&service).Error; err != nil {
		log.InfoWithFields(err.Error(), log.Fields{
			"id": r.ServiceSpec.ServiceID,
		})
		return nil, err
	}

	return &ServiceResolver{
		DB:      r.DB,
		Service: service,
	}, nil
}
