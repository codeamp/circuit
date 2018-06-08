package graphql_resolver

import (
	"github.com/jinzhu/gorm"
)

// Feature Resolver Initializer
type FeatureResolverInitializer struct {
	DB *gorm.DB
}
