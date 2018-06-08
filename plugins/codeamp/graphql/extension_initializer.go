package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// Extension Resolver Initializer
type ExtensionResolverInitializer struct {
	DB *gorm.DB
}
