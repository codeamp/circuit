package resolvers

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionEnvironmentVariableInput struct {
	ProjectId             *string
	Key                   string
	Type                  string
	EnvironmentVariableId *string
}

type ExtensionSpecInput struct {
	ID                   *string
	Name                 string
	Component            string
	Type                 string
	Key                  string
	EnvironmentVariables []*ExtensionEnvironmentVariableInput
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
	// extensionSpec := models.ExtensionSpec{
	// 	Name:      args.ExtensionSpec.Name,
	// 	Component: args.ExtensionSpec.Component,
	// 	Type:      plugins.Type(args.ExtensionSpec.Type),
	// 	Key:       args.ExtensionSpec.Key,

	// }
	// r.db.Create(&extensionSpec)
	// r.actions.ExtensionSpecCreated(&extensionSpec)

	// return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, nil
}

func (r *Resolver) UpdateExtensionSpec(args *struct{ ExtensionSpec *ExtensionSpecInput }) (*ExtensionSpecResolver, error) {
	// marshal args.ExtensionSpec.EnvironmentVariables for JSONB storage in db
	marshalledEnvVars, err := json.Marshal(args.ExtensionSpec.EnvironmentVariables)
	if err != nil {
		log.InfoWithFields("could not marshal args.ExtensionSpec.EnvironmentVariables", log.Fields{
			"v": args.ExtensionSpec.EnvironmentVariables,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, nil
	}

	extensionSpec := models.ExtensionSpec{}
	if r.db.Where("id = ?", args.ExtensionSpec.ID).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("could not find extensionspec with id", log.Fields{
			"id": args.ExtensionSpec.ID,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, nil
	}

	// update extensionspec properties
	extensionSpec.Name = args.ExtensionSpec.Name
	extensionSpec.Key = args.ExtensionSpec.Key
	extensionSpec.Type = plugins.Type(args.ExtensionSpec.Type)
	extensionSpec.Component = args.ExtensionSpec.Component
	extensionSpec.EnvironmentVariables = postgres.Jsonb{marshalledEnvVars}

	r.db.Save(&extensionSpec)

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
	return string(r.ExtensionSpec.Type)
}

func (r *ExtensionSpecResolver) Key() string {
	return r.ExtensionSpec.Key
}

func (r *ExtensionSpecResolver) EnvironmentVariables(ctx context.Context) ([]*EnvironmentVariableResolver, error) {
	var results []*EnvironmentVariableResolver

	return results, nil
}

func (r *ExtensionSpecResolver) FormSpec(ctx context.Context) ([]*KeyValueResolver, error) {
	var keyValues []plugins.KeyValue
	// err := plugins.ConvertMapStringStringToKV(r.ExtensionSpec.FormSpec, &keyValues)
	// if err != nil {
	// 	log.InfoWithFields("not able to convert map[string]string to keyvalues", log.Fields{
	// 		"extensionSpec": r.ExtensionSpec,
	// 	})
	// }

	var results []*KeyValueResolver
	for _, kv := range keyValues {
		results = append(results, &KeyValueResolver{db: r.db, KeyValue: kv})
	}

	return results, nil
}

func (r *ExtensionSpecResolver) Created() graphql.Time {
	return graphql.Time{Time: r.ExtensionSpec.Model.CreatedAt}
}
