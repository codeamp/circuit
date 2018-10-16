package graphql_resolver

import (
	"context"
	"fmt"

	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/jinzhu/gorm"
)

// Environment Resolver Query
type EnvironmentResolverMutation struct {
	DB *gorm.DB
}

func (r *EnvironmentResolverMutation) CreateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv model.Environment
	if r.DB.Where("key = ?", args.Environment.Key).Find(&existingEnv).RecordNotFound() {
		env := model.Environment{
			Name:      args.Environment.Name,
			Key:       args.Environment.Key,
			IsDefault: args.Environment.IsDefault,
			Color:     args.Environment.Color,
		}

		r.DB.Create(&env)

		//r.EnvironmentCreated(&env)

		return &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: env}}, nil
	} else {
		return nil, fmt.Errorf("CreateEnvironment: name '%s' already exists", args.Environment.Name)
	}
}

func (r *EnvironmentResolverMutation) DeleteEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv model.Environment
	if r.DB.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("DeleteEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		// if this is the only default env, do not delete
		if existingEnv.IsDefault {
			var defaultEnvs []model.Environment
			r.DB.Where("is_default = ?", true).Find(&defaultEnvs)
			if len(defaultEnvs) == 1 {
				return nil, fmt.Errorf("Cannot delete since this is the only default env. Must be one at all times")
			}
		}

		// Only delete env. if no child services exist, else return err
		childServices := []model.Service{}
		r.DB.Where("environment_id = ?", args.Environment.ID).Find(&childServices)
		if len(childServices) == 0 {
			existingEnv.Name = args.Environment.Name
			secrets := []model.Secret{}

			r.DB.Delete(&existingEnv)
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Find(&secrets)
			for _, secret := range secrets {
				r.DB.Delete(&secret)
				r.DB.Where("secret_id = ?", secret.Model.ID).Delete(model.SecretValue{})
			}

			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(model.Release{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(model.ProjectExtension{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(model.ProjectSettings{})
			r.DB.Where("environment_id = ?", existingEnv.Model.ID).Delete(model.Extension{})

			//r.EnvironmentDeleted(&existingEnv)

			return &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: existingEnv}}, nil
		} else {
			return nil, fmt.Errorf("Delete all project-services in environment before deleting environment.")
		}
	}
}

func (r *EnvironmentResolverMutation) UpdateEnvironment(ctx context.Context, args *struct{ Environment *model.EnvironmentInput }) (*EnvironmentResolver, error) {
	var existingEnv model.Environment

	if args.Environment.ID == nil {
		return &EnvironmentResolver{}, fmt.Errorf("EnvironmentID required param")
	}

	if r.DB.Where("id = ?", args.Environment.ID).Find(&existingEnv).RecordNotFound() {
		return nil, fmt.Errorf("UpdateEnv: couldn't find environment: %s", *args.Environment.ID)
	} else {
		existingEnv.Name = args.Environment.Name
		existingEnv.Color = args.Environment.Color
		// Check if this is the only default env.
		if args.Environment.IsDefault == false && existingEnv.IsDefault == true {
			var defaultEnvs []model.Environment
			r.DB.Where("is_default = ?", true).Find(&defaultEnvs)
			// Update IsDefault as long as the current is false or
			// if there are more than 1 default env
			if len(defaultEnvs) > 1 {
				existingEnv.IsDefault = args.Environment.IsDefault
			}
		} else {
			// If IsDefault is true, then no harm in updating
			existingEnv.IsDefault = args.Environment.IsDefault
		}

		r.DB.Save(&existingEnv)

		return &EnvironmentResolver{DBEnvironmentResolver: &db_resolver.EnvironmentResolver{DB: r.DB, Environment: existingEnv}}, nil
	}
}
