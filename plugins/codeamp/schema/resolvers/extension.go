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
	val, _ := r.Extension.Artifacts.Value()
	err := plugins.ConvertMapStringStringToKV(val.(map[string]*string), &keyValues)
	if err != nil {
		log.InfoWithFields("not able to convert map[string]string to keyvalues", log.Fields{
			"extensionSpec": r.ExtensionSpec,
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
	/*
		var keyValues []plugins.KeyValue
		val, _ := r.Extension.FormSpecValues.Value()
		err := plugins.ConvertMapStringStringToKV(val.(map[string]*string), &keyValues)
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
	*/
	return []*KeyValueResolver{}, nil
}

func (r *Resolver) CreateExtension(ctx context.Context, args *struct{ Extension *ExtensionInput }) (*ExtensionResolver, error) {
	var extension models.Extension
	var formSpecValuesMap map[string]*string

	// check if extension already exists with project
	if r.db.Where("project_id = ? and extension_spec_id = ?", args.Extension.ProjectId, args.Extension.ExtensionSpecId).Find(&extension).RecordNotFound() {
		// make sure extension form spec values are valid
		// if they are valid, create extension object

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

		err = plugins.ConvertKVToMapStringString(args.Extension.FormSpecValues, &formSpecValuesMap)
		if err != nil {
			log.InfoWithFields("can't convert kv to map[string]*string", log.Fields{
				"extension": args.Extension,
			})
			return nil, err
		}

		extension = models.Extension{
			ExtensionSpecId: extensionSpecId,
			ProjectId:       projectId,
			FormSpecValues:  formSpecValuesMap,
			Artifacts:       map[string]*string{},
			State:           plugins.Waiting,
			Slug:            "",
		}

		r.db.Create(&extension)

		extension.Slug = fmt.Sprintf("dockerbuild%s", extension.Model.ID.String())
		r.db.Save(&extension)

		r.actions.ExtensionCreated(&extension)
	}
	return &ExtensionResolver{db: r.db, Extension: extension}, nil
}
