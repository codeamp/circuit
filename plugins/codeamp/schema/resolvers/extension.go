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

type ExtensionResolver struct {
	db        *gorm.DB
	Extension models.Extension
}

type ExtensionInput struct {
	ID              *string
	ProjectId       string
	ExtensionSpecId string
	FormSpecValues  []plugins.KeyValue
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

	// check if extension already exists with project
	if r.db.Where("project_id = ? and extension_spec_id = ?", args.Extension.ProjectId, args.Extension.ExtensionSpecId).Find(&extension).RecordNotFound() {
		// make sure extension form spec values are valid
		// if they are valid, create extension object

		var extensionSpec models.ExtensionSpec

		extensionSpecId, err := uuid.FromString(args.Extension.ExtensionSpecId)
		if err != nil {
			log.InfoWithFields("couldn't parse ExtensionSpecId", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		projectId, err := uuid.FromString(args.Extension.ProjectId)
		if err != nil {
			log.InfoWithFields("couldn't parse ProjectId", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		// validate from formSpec
		valid, err := FormSpecValuesIsValid(r.db, args.Extension)
		if err != nil {
			return nil, err
		}
		if valid == false {
			log.InfoWithFields("form spec values are invalid", log.Fields{
				"extension": args.Extension,
			})
			return nil, nil
		}

		err = plugins.ConvertKVToMapStringString(args.Extension.FormSpecValues, &formSpecValuesMap)
		if err != nil {
			log.InfoWithFields("can't convert kv to map[string]*string", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		if r.db.Where("id = ?", extensionSpecId).Find(&extensionSpec).RecordNotFound() {
			log.InfoWithFields("can't find corresponding extensionSpec", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		spew.Dump("CREATE EXTENSION")
		spew.Dump(extensionSpec)

		extension = models.Extension{
			ExtensionSpecId: extensionSpecId,
			ProjectId:       projectId,
			FormSpecValues:  formSpecValuesMap,
			Artifacts:       map[string]*string{},
			State:           plugins.Waiting,
			Slug:            "",
		}

		r.db.Create(&extension)

		extension.Slug = fmt.Sprintf("%s|%s", extensionSpec.Key, extension.Model.ID.String())
		r.db.Save(&extension)

		r.actions.ExtensionCreated(&extension)
	}
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
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

func FormSpecValuesIsValid(db *gorm.DB, extensionInput *ExtensionInput) (bool, error) {
	// get extension spec
	var extensionSpec models.ExtensionSpec
	if db.Where("id = ?", extensionInput.ExtensionSpecId).Find(&extensionSpec).RecordNotFound() {
		log.InfoWithFields("extensionSpec not found", log.Fields{
			"extensionInput": extensionInput,
		})
		return false, nil
	}

	// loop through extension spec

	// convert extensionSpec's form spec values into KV array
	var extensionSpecKVFormSpec []plugins.KeyValue
	err := plugins.ConvertMapStringStringToKV(extensionSpec.FormSpec, &extensionSpecKVFormSpec)
	if err != nil {
		return false, err
	}

	// convert extensionInput's form spec values into map[string]string
	extensionInputMap := make(map[string]*string)
	err = plugins.ConvertKVToMapStringString(extensionInput.FormSpecValues, &extensionInputMap)
	if err != nil {
		return false, err
	}

	for _, kv := range extensionSpecKVFormSpec {
		if extensionInputMap[kv.Key] == nil {
			return false, nil
		}
	}

	return true, nil
}
