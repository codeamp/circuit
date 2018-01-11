package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
)

type EnvironmentVariableValueResolver struct {
	db               *gorm.DB
	EnvironmentVariableValue models.EnvironmentVariableValue
}

func (r *Resolver) EnvironmentVariableValue(ctx context.Context, args *struct{ ID graphql.ID }) (*EnvironmentVariableValueResolver, error) {
	environmentVariableValue := models.EnvironmentVariableValue{}
	if err := r.db.Where("id = ?", args.ID).First(&environmentVariableValue).Error; err != nil {
		return nil, err
	}

	return &EnvironmentVariableValueResolver{db: r.db, EnvironmentVariableValue: environmentVariableValue}, nil
}

func (r *EnvironmentVariableValueResolver) ID() graphql.ID {
	return graphql.ID(r.EnvironmentVariableValue.Model.ID.String())
}

func (r *EnvironmentVariableValueResolver) Value() string {
	return r.EnvironmentVariableValue.Value
}

func (r *EnvironmentVariableValueResolver) EnvironmentVariable(ctx context.Context) (*EnvironmentVariableResolver, error) {
	envVar := models.EnvironmentVariable{}
	if r.db.Where("id = ?", r.EnvironmentVariableValue.EnvironmentVariableId.String()).Find(&envVar).RecordNotFound() {
		log.InfoWithFields("envvar not found", log.Fields{
			"id": r.EnvironmentVariableValue.EnvironmentVariableId.String(),
		})
		return nil, fmt.Errorf("envvar not found")
	}
	return &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar}, nil
}

func (r *EnvironmentVariableValueResolver) User(ctx context.Context) (*UserResolver, error) {
	user := models.User{}
	if r.db.Where("id = ?", r.EnvironmentVariableValue.UserId.String()).Find(&user).RecordNotFound() {
		log.InfoWithFields("user not found", log.Fields{
			"id": r.EnvironmentVariableValue.UserId.String(),
		})
		return &UserResolver{db: r.db, User: user}, fmt.Errorf("user not found")
	}
	return &UserResolver{db: r.db, User: user}, nil
}

func (r *EnvironmentVariableValueResolver) Created() graphql.Time {
	return graphql.Time{Time: r.EnvironmentVariableValue.Model.CreatedAt}
}
