package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ServiceSpecResolver struct {
	model.ServiceSpec
	DB *gorm.DB
}

func (r *ServiceSpecResolver) Service() *ServiceResolver {
	service := model.Service{}
	if err := r.DB.Where("id = ?", r.ServiceSpec.ServiceID).First(&service).Error; err != nil {
		return nil
	}
	
	return &ServiceResolver{
		DB: r.DB,
		Service: service,
	}
}