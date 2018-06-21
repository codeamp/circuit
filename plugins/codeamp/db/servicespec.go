package db_resolver

import (
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

type ServiceSpecResolver struct {
	model.ServiceSpec
	DB *gorm.DB
}
