package resolvers

import (
	"context"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type ExtensionSpecEnvironmentVariableInput struct {
	Type                string
	EnvironmentVariable *string
	Key                 string
}

func (r *Resolver) ExtensionSpecEnvironmentVariable(ctx context.Context, args *struct{ ID graphql.ID }) *ExtensionSpecEnvironmentVariableResolver {
	extensionSpecEnvironmentVariable := models.ExtensionSpecEnvironmentVariable{}
	return &ExtensionSpecEnvironmentVariableResolver{db: r.db, ExtensionSpecEnvironmentVariable: extensionSpecEnvironmentVariable}
}

type ExtensionSpecEnvironmentVariableResolver struct {
	db                               *gorm.DB
	ExtensionSpecEnvironmentVariable models.ExtensionSpecEnvironmentVariable
}

func (r *ExtensionSpecEnvironmentVariableResolver) ID() graphql.ID {
	return graphql.ID(r.ExtensionSpecEnvironmentVariable.Model.ID.String())
}

func (r *ExtensionSpecEnvironmentVariableResolver) ExtensionSpec(ctx context.Context) (*ExtensionSpecResolver, error) {
	extensionSpec := models.ExtensionSpec{}

	if r.db.Where("id = ?", r.ExtensionSpecEnvironmentVariable.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extension spec not found", log.Fields{
			"id": r.ExtensionSpecEnvironmentVariable.ExtensionSpecId,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
	}

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *ExtensionSpecEnvironmentVariableResolver) EnvironmentVariable(ctx context.Context) (*EnvironmentVariableResolver, error) {
	environmentVariable := models.EnvironmentVariable{}

	if r.db.Where("id = ?", r.ExtensionSpecEnvironmentVariable.EnvironmentVariableId).Find(&environmentVariable).RecordNotFound() {
		log.InfoWithFields("environment variable not found", log.Fields{
			"id": r.ExtensionSpecEnvironmentVariable.EnvironmentVariableId,
		})
		return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: environmentVariable}, nil
	}

	return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: environmentVariable}, nil
}

func (r *ExtensionSpecEnvironmentVariableResolver) Type() string {
	return string(r.ExtensionSpecEnvironmentVariable.Type)
}

func (r *ExtensionSpecEnvironmentVariableResolver) Key() string {
	return r.ExtensionSpecEnvironmentVariable.Key
}

func (r *ExtensionSpecEnvironmentVariableResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpecEnvironmentVariable.Model.CreatedAt}
}
