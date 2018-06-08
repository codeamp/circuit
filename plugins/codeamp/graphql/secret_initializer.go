package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// Secret Resolver Initializer
type SecretResolverInitializer struct {
	DB *gorm.DB
}
