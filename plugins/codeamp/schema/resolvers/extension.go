package resolvers

import (
	"context"
	"errors"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	graphql "github.com/neelance/graphql-go"
	uuid "github.com/satori/go.uuid"
)

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
}

func NewExtensionResolver(extension models.Extension, db *gorm.DB) *ExtensionResolver {
	return &ExtensionResolver{
		db:        db,
		Extension: extension,
	}
}

type ExtensionInput struct {
	ID              *string
	ProjectId       string
	ExtensionSpecId string
	FormSpecValues  []plugins.KeyValue
	EnvironmentId   string
}

func (r *Resolver) Extension(ctx context.Context, args *struct{ ID graphql.ID }) (*ExtensionResolver, error) {
	extension := models.Extension{}
	if err := r.db.Where("id = ?", args.ID).First(&extension).Error; err != nil {
		return nil, err
	}

	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ExtensionResolver) ID() graphql.ID {
	return graphql.ID(r.Extension.Model.ID.String())
}

func (r *ExtensionResolver) Project(ctx context.Context) (*ProjectResolver, error) {
	var project models.Project
	r.db.Model(r.Extension).Related(&project)
	return &ProjectResolver{db: r.db, Project: project}, nil
}

func (r *ExtensionResolver) ExtensionSpec(ctx context.Context) (*ExtensionSpecResolver, error) {
	var extensionSpec models.ExtensionSpec
	r.db.Model(r.Extension).Related(&extensionSpec)
	return &ExtensionSpecResolver{db: r.db, ExtensionSpec: extensionSpec}, nil
}

func (r *ExtensionResolver) State() string {
	return string(r.Extension.State)
}

func (r *ExtensionResolver) Artifacts() []*KeyValueResolver {
	var keyValues []plugins.KeyValue
	err := plugins.ConvertMapStringStringToKV(r.Extension.Artifacts, &keyValues)
	if err != nil {
		log.InfoWithFields("not able to convert map[string]string to keyvalues", log.Fields{
			"extensionSpec": r.Extension,
		})
	}
	var rows []*KeyValueResolver
	for _, kv := range keyValues {
		rows = append(rows, &KeyValueResolver{db: r.db, KeyValue: kv})
	}
	return rows
}

func (r *ExtensionResolver) Created() graphql.Time {
	return graphql.Time{Time: r.Extension.Model.CreatedAt}
}

func (r *ExtensionResolver) FormSpecValues(ctx context.Context) ([]*KeyValueResolver, error) {
	var keyValues []plugins.KeyValue
	err := plugins.ConvertMapStringStringToKV(r.Extension.FormSpecValues, &keyValues)
	if err != nil {
		log.InfoWithFields("not able to convert map[string]string to keyvalues", log.Fields{
			"extensionSpec": r.Extension,
		})
		return nil, err
	}

	var rows []*KeyValueResolver
	for _, kv := range keyValues {
		rows = append(rows, &KeyValueResolver{db: r.db, KeyValue: kv})
	}

	return rows, nil
}

func (r *Resolver) CreateExtension(ctx context.Context, args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	formSpecValuesMap := make(map[string]*string)

	extensionSpecId, err := uuid.FromString(args.Extension.ExtensionSpecId)
	if err != nil {
		log.InfoWithFields("couldn't parse ExtensionSpecId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse ExtensionSpecId. Invalid Format.")
	}

	projectId, err := uuid.FromString(args.Extension.ProjectId)
	if err != nil {
		log.InfoWithFields("couldn't parse ProjectId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse ProjectId. Invalid format.")
	}

	environmentId, err := uuid.FromString(args.Extension.EnvironmentId)
	if err != nil {
		log.InfoWithFields("couldn't parse EnvironmentId", log.Fields{
			"extension": args.Extension,
		})
		return nil, errors.New("Could not parse EnvironmentId. Invalid format.")
	}

	// check if extension already exists with project
	if r.db.Where("project_id = ? and extension_spec_id = ?", projectId, extensionSpecId).Find(&extension).RecordNotFound() {
		// make sure extension form spec values are valid
		// if they are valid, create extension object

		var extensionSpec models.ExtensionSpec

		if r.db.Where("id = ?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("can't find corresponding extensionSpec", log.Fields{
				"extension": args.Extension,
			})
			return nil, errors.New("Can't find corresponding extensionSpec.")
		}

		err = plugins.ConvertKVToMapStringString(args.Extension.FormSpecValues, &formSpecValuesMap)
		if err != nil {
			log.InfoWithFields("can't convert kv to map[string]*string", log.Fields{
				"extension": args.Extension,
			})
			return nil, errors.New("Can't convert kv to map[string]*string")
		}

		// validate from formSpec
		err := FormSpecValuesIsValid(r.db, args.Extension)
		if err != nil {
			log.InfoWithFields("FormSpecValuesIsValid failed", log.Fields{
				"err": err,
			})
			return nil, err
		}

		extension = models.Extension{
			ExtensionSpecId: extensionSpecId,
			ProjectId:       projectId,
			EnvironmentId:   environmentId,
			FormSpecValues:  formSpecValuesMap,
			Artifacts:       map[string]*string{},
			State:           plugins.Waiting,
			Slug:            "",
		}

		r.db.Create(&extension)

		extension.Slug = fmt.Sprintf("%s|%s", extensionSpec.Key, extension.Model.ID.String())
		r.db.Save(&extension)

		go r.actions.ExtensionCreated(&extension)
		return &ExtensionResolver{db: r.db, Extension: extension}, nil
	}
	return nil, errors.New("This extension is already installed in this project.")
}

func (r *Resolver) UpdateExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	formSpecValuesMap := make(map[string]*string)

	if r.db.Where("id = ?", args.Extension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}

	err := plugins.ConvertKVToMapStringString(args.Extension.FormSpecValues, &formSpecValuesMap)
	if err != nil {
		log.InfoWithFields("not able to convert kv to map[string]string", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}

	extension.FormSpecValues = formSpecValuesMap
	extension.State = plugins.Waiting

	r.db.Save(&extension)
	r.actions.ExtensionUpdated(&extension)

	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *Resolver) DeleteExtension(args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	var res []models.ReleaseExtension

	if r.db.Where("id = ?", args.Extension.ID).First(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"extension": args.Extension,
		})
		return &ExtensionResolver{}, nil
	}

	// delete all release extension objects with extension id
	if r.db.Where("extension_id = ?", args.Extension.ID).Find(&res).RecordNotFound() {
		log.InfoWithFields("no release extensions found", log.Fields{
			"extension": extension,
		})
		return &ExtensionResolver{}, nil
	}

	for _, re := range res {
		r.db.Delete(&re)
	}

	r.db.Delete(&extension)
	r.actions.ExtensionDeleted(&extension)
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}

func (r *ExtensionResolver) Environment(ctx context.Context) (*EnvironmentResolver, error) {
	var environment models.Environment
	if r.db.Where("id = ?", r.Extension.EnvironmentId).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"service": r.Extension,
		})
		return nil, fmt.Errorf("Environment not found.")
	}
	return &EnvironmentResolver{db: r.db, Environment: environment}, nil
}

func FormSpecValuesIsValid(db *gorm.DB, extensionInput *ExtensionInput) error {
	// get extension spec
	var extensionSpec models.ExtensionSpec
	var missingKeys []string

	if db.Where("id = ?", extensionInput.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extensionInput": extensionInput,
		})
		return errors.New("ExtensionSpec not found")
	}

	// loop through extension spec

	// convert extensionSpec's form spec values into KV array
	var extensionSpecKVFormSpec []plugins.KeyValue
	err := plugins.ConvertMapStringStringToKV(extensionSpec.FormSpec, &extensionSpecKVFormSpec)
	if err != nil {
		return err
	}

	// convert extensionInput's form spec values into map[string]string
	extensionInputMap := make(map[string]*string)
	err = plugins.ConvertKVToMapStringString(extensionInput.FormSpecValues, &extensionInputMap)
	if err != nil {
		return err
	}

	for _, kv := range extensionSpecKVFormSpec {
		if extensionInputMap[kv.Key] == nil {
			missingKeys = append(missingKeys, kv.Key)
		}
	}

	if len(missingKeys) > 0 {
		return fmt.Errorf("Required keys not found within extension input: %v", missingKeys)
	}

	return nil
}
