package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// Environment Resolver Initializer
type EnvironmentResolverInitializer struct {
	DB *gorm.DB
}
