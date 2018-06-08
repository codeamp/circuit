package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// Service Resolver Initializer
type ServiceResolverInitializer struct {
	DB *gorm.DB
}
