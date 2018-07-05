package graphql_resolver

import (
	"github.com/codeamp/transistor"
	"github.com/go-redis/redis"
	"github.com/jinzhu/gorm"
)

// Resolver is the main resolver for all queries
type Resolver struct {
	// DB
	DB *gorm.DB
	// Events
	Events chan transistor.Event
	// Redis
	Redis *redis.Client
}
