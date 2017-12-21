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

type ExtensionSpecInput struct {
	ID            *string
	Name          string
	Component     string
	Type          string
	Key           string
	EnvironmentId string
	Config        string
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
	// convert to object from KV
	convertedConfig := make(map[string]interface{})
	err := json.Unmarshal([]byte(args.ExtensionSpec.Config), &convertedConfig)
	if err != nil {
		log.InfoWithFields(err.Error(), log.Fields{
			"v": args.ExtensionSpec.Config,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, err
	}

	// marshal args.ExtensionSpec.Configs for JSONB storage in db
	marshalledConfig, err := json.Marshal(convertedConfig)
	if err != nil {
		log.InfoWithFields("could not marshal convertedConfig", log.Fields{
			"v": convertedConfig,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, err
	}

	extensionSpec := models.ExtensionSpec{}
	if r.db.Where("id = ?", args.ExtensionSpec.ID).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("could not find extensionspec with id", log.Fields{
			"id": args.ExtensionSpec.ID,
		})
		return &ExtensionSpecResolver{db: r.db, ExtensionSpec: models.ExtensionSpec{}}, fmt.Errorf("could not find extensionspec with id")
	}

	environmentId, err := uuid.FromString(args.ExtensionSpec.EnvironmentId)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentId")
	}

	// update extensionspec properties
	extensionSpec.Name = args.ExtensionSpec.Name
	extensionSpec.Key = args.ExtensionSpec.Key
	extensionSpec.Type = plugins.Type(args.ExtensionSpec.Type)
	extensionSpec.Component = args.ExtensionSpec.Component
	extensionSpec.EnvironmentId = environmentId
	extensionSpec.Config = postgres.Jsonb{marshalledConfig}

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

func (r *ExtensionSpecResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	environment := models.Environment{}

	if r.db.Where("id = ?", r.ExtensionSpec.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": r.ExtensionSpec.EnvironmentId,
		})
		return nil, fmt.Errorf("Environment not found.")
	}

	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}

func (r *ExtensionSpecResolver) Config(ctx context.Context) (string, error) {
	tmp := make(map[string]interface{})
	err := json.Unmarshal(r.ExtensionSpec.Config.RawMessage, &tmp)
	if err != nil {
		log.InfoWithFields("could not unmarshal r.ExtensionSpec.Configs", log.Fields{
			"data": r.ExtensionSpec.Config,
			"v":    &tmp,
		})
		return "", err
	}

	str, err := json.Marshal(tmp)
	if err != nil {
		log.Info("could not marshal config")
		return "", err
	}

	return string(str), nil
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
