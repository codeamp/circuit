package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/davecgh/go-spew/spew"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionSpecInput struct {
	ID                   *string
	Name                 string
	Component            string
	EnvironmentVariables []*ExtensionSpecEnvironmentVariableInput
	Type                 string
	Key                  string
}

func (r *Resolver) ExtensionSpec(ctx context.Context, args *struct{ ID graphql.ID }) *ExtensionSpecResolver {
	extensionSpec := models.ExtensionSpec{}
	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}
}

type ExtensionSpecResolver struct {
	db            *gorm.DB
	ExtensionSpec models.ExtensionSpec
}

func (r *Resolver) CreateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {

	// Check if valid plugins.ExtensionSpecType
	envVarType := plugins.StrToExtensionType[args.ExtensionSpec.Type]
	if envVarType == "" {
		return &ExtensionSpecResolver{}, fmt.Errorf("Invalid extension type: %s", args.ExtensionSpec.Type)
	}

	extensionSpec := models.ExtensionSpec{
		Name:      args.ExtensionSpec.Name,
		Component: args.ExtensionSpec.Component,
		Type:      plugins.ExtensionType(args.ExtensionSpec.Type),
		Key:       args.ExtensionSpec.Key,
	}

	r.db.Create(&extensionSpec)

	spew.Dump(args.ExtensionSpec)

	// create extension spec env vars
	for _, envVar := range args.ExtensionSpec.EnvironmentVariables {
		envVarId := uuid.FromStringOrNil(*envVar.EnvironmentVariable)

		// check if env var exists
		var dbEnvVar models.EnvironmentVariable
		if r.db.Where("id = ?", envVarId).Find(&dbEnvVar).RecordNotFound() {
			return &ExtensionSpecResolver{}, fmt.Errorf("Specified env vars don't exist.")
		}
		spew.Dump("EXTENSION SPEC ENV VAR CREATING!")
		extensionSpecEnvVar := models.ExtensionSpecEnvironmentVariable{
			ExtensionSpecId:       extensionSpec.Model.ID,
			EnvironmentVariableId: envVarId,
			Type: plugins.ExtensionSpecEnvVarType(envVar.Type),
			Key:  envVar.Key,
		}
		spew.Dump("CEATED!")
		r.db.Create(&extensionSpecEnvVar)
	}

	r.actions.ExtensionSpecCreated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) UpdateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	var extensionSpec models.ExtensionSpec

	if r.db.Where("id = ?", args.ExtensionSpec.ID).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ExtensionSpec,
		})
		return &ExtensionSpecResolver{}, nil
	}

	extensionSpec.Name = args.ExtensionSpec.Name
	extensionSpec.Component = args.ExtensionSpec.Component
	extensionSpec.Key = args.ExtensionSpec.Key
	extensionSpec.Type = plugins.ExtensionType(args.ExtensionSpec.Type)

	r.db.Save(&extensionSpec)

	// delete all old extension spec env vars
	r.db.Where("extension_spec_id = ?", extensionSpec.Model.ID.String()).Delete(&models.ExtensionSpecEnvironmentVariable{})

	// create extension spec env vars
	for _, envVar := range args.ExtensionSpec.EnvironmentVariables {
		envVarId := uuid.FromStringOrNil(*envVar.EnvironmentVariable)
		extensionSpecEnvVar := models.ExtensionSpecEnvironmentVariable{
			ExtensionSpecId:       extensionSpec.Model.ID,
			EnvironmentVariableId: envVarId,
			Type: plugins.ExtensionSpecEnvVarType(envVar.Type),
			Key:  envVar.Key,
		}
		r.db.Save(&extensionSpecEnvVar)
	}

	r.actions.ExtensionSpecUpdated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) DeleteExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	extensionSpec := models.ExtensionSpec{}
	extensions := []models.Extension{}

	extensionSpecId, err := uuid.FromString(*args.ExtensionSpec.ID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	if r.db.Where("id=?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
		return nil, fmt.Errorf("ExtensionSpec not found with given argument id")
	}

	// delete all extensions using extension spec
	if r.db.Where("extension_spec_id = ?", extensionSpecId).Find(&extensions).RecordNotFound() {
		log.InfoWithFields("no extensions using this extension spec", log.Fields{
			"extension spec": extensionSpec,
		})
	}

	if len(extensions) > 0 {
		return nil, fmt.Errorf("You must delete all extensions using this extension spec in order to delete this extension spec.")
	} else {
		r.db.Delete(&extensionSpec)
		r.actions.ExtensionSpecDeleted(&extensionSpec)

		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
	}
}

func (r *ExtensionSpecResolver) ID() graphql.ID {
	return graphql.ID(r.ExtensionSpec.Model.ID.String())
}

func (r *ExtensionSpecResolver) Name() string {
	return r.ExtensionSpec.Name
}

func (r *ExtensionSpecResolver) Component() string {
	return r.ExtensionSpec.Component
}

func (r *ExtensionSpecResolver) Type() string {
	return string(r.ExtensionSpec.Type)
}

func (r *ExtensionSpecResolver) Key() string {
	return r.ExtensionSpec.Key
}

func (r *ExtensionSpecResolver) EnvironmentVariables(ctx context.Context) ([]*ExtensionSpecEnvironmentVariableResolver, error) {

	var extensionSpecEnvVarRows []models.ExtensionSpecEnvironmentVariable
	var results []*ExtensionSpecEnvironmentVariableResolver

	if r.db.Where("extension_spec_id = ?", r.ExtensionSpec.Model.ID.String()).Find(&extensionSpecEnvVarRows).RecordNotFound() {
		log.InfoWithFields("no extension spec env vars found", log.Fields{
			"extension_spec_id": r.ExtensionSpec.Model.ID.String(),
		})
		return []*ExtensionSpecEnvironmentVariableResolver{}, nil
	}

	for _, extensionSpecEnvVar := range extensionSpecEnvVarRows {
		results = append(results, &ExtensionSpecEnvironmentVariableResolver{db: r.db, ExtensionSpecEnvironmentVariable: extensionSpecEnvVar})
	}

	return results, nil
}

func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Model.CreatedAt}
}
