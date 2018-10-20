package graphql_resolver

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
	uuid "github.com/satori/go.uuid"
)

// Environment Resolver Query
type ExtensionResolverMutation struct {
	DB *gorm.DB
}

func (r *ExtensionResolverMutation) CreateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	environmentID, err := uuid.FromString(args.Extension.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentID")
	}

	ext := model.Extension{
		Name:          args.Extension.Name,
		Component:     args.Extension.Component,
		Type:          plugins.Type(args.Extension.Type),
		Key:           args.Extension.Key,
		EnvironmentID: environmentID,
		Config:        postgres.Jsonb{[]byte(args.Extension.Config.RawMessage)},
	}

	r.DB.Create(&ext)
	//r.ExtensionCreated(&ext)

	return &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: ext}}, nil
}

func (r *ExtensionResolverMutation) UpdateExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	ext := model.Extension{}
	if r.DB.Where("id = ?", args.Extension.ID).Find(&ext).RecordNotFound() {
		log.InfoWithFields("could not find extensionspec with id", log.Fields{
			"id": args.Extension.ID,
		})
		return &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: model.Extension{}}}, fmt.Errorf("could not find extensionspec with id")
	}

	environmentID, err := uuid.FromString(args.Extension.EnvironmentID)
	if err != nil {
		return nil, fmt.Errorf("Missing argument EnvironmentID")
	}

	// update extensionspec properties
	ext.Name = args.Extension.Name
	ext.Key = args.Extension.Key
	ext.Type = plugins.Type(args.Extension.Type)
	ext.Component = args.Extension.Component
	ext.EnvironmentID = environmentID
	ext.Config = postgres.Jsonb{args.Extension.Config.RawMessage}

	r.DB.Save(&ext)

	//r.ExtensionUpdated(&ext)

	return &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: ext}}, nil
}

func (r *ExtensionResolverMutation) DeleteExtension(args *struct{ Extension *model.ExtensionInput }) (*ExtensionResolver, error) {
	ext := model.Extension{}
	extensions := []model.ProjectExtension{}
	if args.Extension.ID == nil {
		return nil, fmt.Errorf("Missing argument id")
	}

	extID, err := uuid.FromString(*args.Extension.ID)
	if err != nil {
		return nil, fmt.Errorf("Invalid argument id")
	}

	if r.DB.Where("id=?", extID).Find(&ext).RecordNotFound() {
		return nil, fmt.Errorf("Extension not found with given argument id")
	}

	// delete all extensions using extension spec
	if r.DB.Where("extension_id = ?", extID).Find(&extensions).RecordNotFound() {
		log.InfoWithFields("no extensions using this extension spec", log.Fields{
			"extension spec": ext,
		})
	}

	if len(extensions) > 0 {
		return nil, fmt.Errorf("You must delete all extensions using this extension spec in order to delete this extension spec.")
	} else {
		r.DB.Delete(&ext)

		//r.ExtensionDeleted(&ext)

		return &ExtensionResolver{DBExtensionResolver: &db_resolver.ExtensionResolver{DB: r.DB, Extension: ext}}, nil
	}
}
