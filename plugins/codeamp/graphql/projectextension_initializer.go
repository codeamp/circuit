package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// ProjectExtension Resolver Initializer
type ProjectExtensionResolverInitializer struct {
	DB *gorm.DB
}
