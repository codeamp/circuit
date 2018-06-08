package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// ServiceSpec Resolver Initializer
type ServiceSpecResolverInitializer struct {
	DB *gorm.DB
}
