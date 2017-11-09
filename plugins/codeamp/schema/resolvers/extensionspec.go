package resolvers

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionSpecInput struct {
	ID                   *string
	Name                 string
	Component            string
	FormSpec             []plugins.KeyValue
	EnvironmentVariables []map[string]interface{}
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

	formSpecMap := make(map[string]*string)

	err := plugins.ConvertKVToMapStringString(args.ExtensionSpec.FormSpec, &formSpecMap)
	if err != nil {
		return &ExtensionSpecResolver{}, err
	}

	extensionSpec := models.ExtensionSpec{
		Name:      args.ExtensionSpec.Name,
		Component: args.ExtensionSpec.Component,
		FormSpec:  formSpecMap,
		Type:      args.ExtensionSpec.Type,
		Key:       args.ExtensionSpec.Key,
	}

	r.db.Create(&extensionSpec)

	// create extension spec env vars
	for _, envVar := range args.ExtensionSpec.EnvironmentVariables {
		envVarId := uuid.FromStringOrNil(envVar["envVar"].(string))
		extensionSpecEnvVar := models.ExtensionSpecEnvironmentVariable{
			ExtensionSpecId:       extensionSpec.Model.ID,
			EnvironmentVariableId: envVarId,
		}
		r.db.Save(&extensionSpecEnvVar)
	}

	r.actions.ExtensionSpecCreated(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *Resolver) UpdateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	var extensionSpec models.ExtensionSpec
	formSpecMap := make(map[string]*string)

	if r.db.Where("id = ?", args.ExtensionSpec.ID).First(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.ExtensionSpec,
		})
		return &ExtensionSpecResolver{}, nil
	}

	err := plugins.ConvertKVToMapStringString(args.ExtensionSpec.FormSpec, &formSpecMap)
	if err != nil {
		log.InfoWithFields("not able to convert kv to map[string]string", log.Fields{
			"extension": args.ExtensionSpec,
		})
		return &ExtensionSpecResolver{}, nil
	}

	extensionSpec.Name = args.ExtensionSpec.Name
	extensionSpec.Component = args.ExtensionSpec.Component
	extensionSpec.FormSpec = formSpecMap
	extensionSpec.Key = args.ExtensionSpec.Key
	extensionSpec.Type = args.ExtensionSpec.Type

	r.db.Save(&extensionSpec)

	// delete all old extension spec env vars
	r.db.Where("extension_spec_id = ?", extensionSpec.Model.ID.String()).Delete(&models.ExtensionSpecEnvironmentVariable{})

	// create extension spec env vars
	for _, envVar := range args.ExtensionSpec.EnvironmentVariables {
		envVarId := uuid.FromStringOrNil(envVar["envVar"].(string))
		extensionSpecEnvVar := models.ExtensionSpecEnvironmentVariable{
			ExtensionSpecId:       extensionSpec.Model.ID,
			EnvironmentVariableId: envVarId,
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

	// delete all release extensions using each extension
	for _, extension := range extensions {
		res := []models.ReleaseExtension{}
		if r.db.Where("extension_id = ?", extension.Model.ID.String()).Find(&res).RecordNotFound() {
			log.InfoWithFields("no release extensions using this extension id", log.Fields{
				"extension": extension,
			})
		}

		r.db.Delete(&res)
		r.db.Delete(&extension)
	}
	r.db.Delete(&extensionSpec)
	r.actions.ExtensionSpecDeleted(&extensionSpec)

	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
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
	return r.ExtensionSpec.Type
}

func (r *ExtensionSpecResolver) Key() string {
	return r.ExtensionSpec.Key
}

func (r *ExtensionSpecResolver) EnvironmentVariables(ctx context.Context) ([]*EnvironmentVariableResolver, error) {

	var extensionSpecEnvVarRows []models.ExtensionSpecEnvironmentVariable
	var results []*EnvironmentVariableResolver

	if r.db.Where("extension_spec_id = ?", r.ExtensionSpec.Model.ID.String()).Find(&extensionSpecEnvVarRows).RecordNotFound() {
		log.InfoWithFields("no extension spec env vars found", log.Fields{
			"extensionspec": r.ExtensionSpec,
		})
		return []*EnvironmentVariableResolver{}, nil
	}

	for _, extensionSpecEnvVar := range extensionSpecEnvVarRows {
		var envVar models.EnvironmentVariable
		if r.db.Where("id = ?", extensionSpecEnvVar.EnvironmentVariableId).Find(&envVar).RecordNotFound() {
			log.InfoWithFields("no env vars found", log.Fields{
				"extensionspec": r.ExtensionSpec,
			})
			return []*EnvironmentVariableResolver{}, nil
		}

		results = append(results, &EnvironmentVariableResolver{db: r.db, EnvironmentVariable: envVar})
	}

	return results, nil
}

func (r *ExtensionSpecResolver) FormSpec(ctx context.Context) ([]*KeyValueResolver, error) {
	var keyValues []plugins.KeyValue
	err := plugins.ConvertMapStringStringToKV(r.ExtensionSpec.FormSpec, &keyValues)
	if err != nil {
		log.InfoWithFields("not able to convert map[string]string to keyvalues", log.Fields{
			"extensionSpec": r.ExtensionSpec,
		})
	}

	var results []*KeyValueResolver
	for _, kv := range keyValues {
		results = append(results, &KeyValueResolver{db: r.db, KeyValue: kv})
	}

	return results, nil
}

func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Created}
}
