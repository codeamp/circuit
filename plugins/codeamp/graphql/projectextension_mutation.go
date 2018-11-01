package graphql_resolver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/auth"
	db_resolver "github.com/codeamp/circuit/plugins/codeamp/db"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

// ProjectExtension Resolver Mutation
type ProjectExtensionResolverMutation struct {
	DB *gorm.DB
	// Events
	Events chan transistor.Event
}

func (r *ProjectExtensionResolverMutation) CreateProjectExtension(ctx context.Context, args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if _, err := auth.CheckAuth(ctx, []string{}); err != nil {
		return nil, err
	}

	// Check if project can create project extension in environment
	if err := r.DB.Where("environment_id = ? and project_id = ?", args.ProjectExtension.EnvironmentID, args.ProjectExtension.ProjectID).Find(&model.ProjectEnvironment{}).Error; err != nil {
		return nil, errors.New("Project not allowed to install extensions in given environment")
	}

	extension := model.Extension{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).Error; err != nil {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, fmt.Errorf("No extension found for id: '%s'", args.ProjectExtension.ExtensionID)
	}

	project := model.Project{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).Error; err != nil {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, fmt.Errorf("No project found: '%s'", args.ProjectExtension.ProjectID)
	}

	env := model.Environment{}
	if err := r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).Error; err != nil {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("No environment found: '%s'", args.ProjectExtension.ProjectID)
	}

	// check if extension already exists with project
	// ignore if the extension type is 'once' (installable many times)
	if extension.Type == plugins.GetType("once") || extension.Type == plugins.GetType("notification") || r.DB.Where("project_id = ? and extension_id = ? and environment_id = ?", args.ProjectExtension.ProjectID, args.ProjectExtension.ExtensionID, args.ProjectExtension.EnvironmentID).Find(&projectExtension).RecordNotFound() {
		if extension.Key == "route53" {
			err := r.handleExtensionRoute53(args, &projectExtension)
			if err != nil {
				return &ProjectExtensionResolver{}, err
			}
		}

		projectExtension = model.ProjectExtension{
			State:         transistor.GetState("waiting"),
			ExtensionID:   extension.Model.ID,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			Config:        postgres.Jsonb{[]byte(args.ProjectExtension.Config.RawMessage)},
			CustomConfig:  postgres.Jsonb{[]byte(args.ProjectExtension.CustomConfig.RawMessage)},
		}

		r.DB.Save(&projectExtension)

		artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}

		projectExtensionEvent := plugins.ProjectExtension{
			ID: projectExtension.Model.ID.String(),
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Slug:       project.Slug,
				Repository: project.Repository,
			},
			Environment: env.Key,
		}
		ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("create"), projectExtensionEvent)
		ev.Artifacts = artifacts
		r.Events <- ev

		return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
	}

	return nil, errors.New("This extension is already installed in this project.")
}

func (r *ProjectExtensionResolverMutation) UpdateProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no project extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, fmt.Errorf("No project extension found")
	}

	extension := model.Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, fmt.Errorf("No extension found.")
	}

	project := model.Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, fmt.Errorf("No project found.")
	}

	env := model.Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, fmt.Errorf("No environment found.")
	}

	if extension.Key == "route53" {
		err := r.handleExtensionRoute53(args, &projectExtension)
		if err != nil {
			return nil, err
		}
	}

	projectExtension.Config = postgres.Jsonb{args.ProjectExtension.Config.RawMessage}
	projectExtension.CustomConfig = postgres.Jsonb{args.ProjectExtension.CustomConfig.RawMessage}
	projectExtension.State = transistor.GetState("waiting")
	projectExtension.StateMessage = ""

	r.DB.Save(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID: projectExtension.Model.ID.String(),
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}

	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("update"), projectExtensionEvent)
	ev.Artifacts = artifacts

	r.Events <- ev

	return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
}

func (r *ProjectExtensionResolverMutation) DeleteProjectExtension(args *struct{ ProjectExtension *model.ProjectExtensionInput }) (*ProjectExtensionResolver, error) {
	var projectExtension model.ProjectExtension

	if r.DB.Where("id = ?", args.ProjectExtension.ID).First(&projectExtension).RecordNotFound() {
		log.InfoWithFields("no project extension found", log.Fields{
			"extension": args.ProjectExtension,
		})
		return nil, fmt.Errorf("No Project Extension Found")
	}

	extension := model.Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return nil, errors.New("No extension found.")
	}

	project := model.Project{}
	if r.DB.Where("id = ?", args.ProjectExtension.ProjectID).Find(&project).RecordNotFound() {
		log.InfoWithFields("no project found", log.Fields{
			"id": args.ProjectExtension.ProjectID,
		})
		return nil, errors.New("No project found.")
	}

	env := model.Environment{}
	if r.DB.Where("id = ?", args.ProjectExtension.EnvironmentID).Find(&env).RecordNotFound() {
		log.InfoWithFields("no env found", log.Fields{
			"id": args.ProjectExtension.EnvironmentID,
		})
		return nil, errors.New("No environment found.")
	}

	// ADB
	// Removed logic here that would delete all existing release extensions associated with this project extension
	// However, that's not really what we want. Doing this means we lose a part of our release history
	// What we really want is to just delete the project extension from future releases and leave the history unaffected

	r.DB.Delete(&projectExtension)

	artifacts, err := ExtractArtifacts(projectExtension, extension, r.DB)
	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	projectExtensionEvent := plugins.ProjectExtension{
		ID: projectExtension.Model.ID.String(),
		Project: plugins.Project{
			ID:         project.Model.ID.String(),
			Slug:       project.Slug,
			Repository: project.Repository,
		},
		Environment: env.Key,
	}
	ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("project:%s", extension.Key)), transistor.GetAction("delete"), projectExtensionEvent)
	ev.Artifacts = artifacts
	r.Events <- ev

	return &ProjectExtensionResolver{DBProjectExtensionResolver: &db_resolver.ProjectExtensionResolver{DB: r.DB, ProjectExtension: projectExtension}}, nil
}

func (r *ProjectExtensionResolverMutation) handleExtensionRoute53(args *struct{ ProjectExtension *model.ProjectExtensionInput }, projectExtension *model.ProjectExtension) error {
	extension := model.Extension{}
	if r.DB.Where("id = ?", args.ProjectExtension.ExtensionID).Find(&extension).RecordNotFound() {
		log.InfoWithFields("no extension found", log.Fields{
			"id": args.ProjectExtension.ExtensionID,
		})
		return errors.New("No extension found.")
	}	

	// HOTFIX: check for existing subdomains for route53
	unmarshaledCustomConfig := make(map[string]interface{})
	err := json.Unmarshal(args.ProjectExtension.CustomConfig.RawMessage, &unmarshaledCustomConfig)
	if err != nil {
		return fmt.Errorf("Could not unmarshal custom config")
	}

	artifacts, err := ExtractArtifacts(*projectExtension, extension, r.DB)
	if err != nil {
		return err
	}

	hostedZoneId := ""
	for _, artifact := range artifacts {
		if artifact.Key == "HOSTED_ZONE_ID" {
			hostedZoneId = strings.ToUpper(artifact.Value.(string))
			break
		}
	}

	existingProjectExtensions := GetProjectExtensionsWithRoute53Subdomain(strings.ToUpper(unmarshaledCustomConfig["subdomain"].(string)), r.DB)
	for _, existingProjectExtension := range existingProjectExtensions {
		if existingProjectExtension.Model.ID.String() != "" {
			// check if HOSTED_ZONE_ID is the same
			var tmpExtension model.Extension

			r.DB.Where("id = ?", existingProjectExtension.ExtensionID).First(&tmpExtension)

			tmpExtensionArtifacts, err := ExtractArtifacts(existingProjectExtension, tmpExtension, r.DB)
			if err != nil {
				return err
			}

			for _, artifact := range tmpExtensionArtifacts {
				if artifact.Key == "HOSTED_ZONE_ID" &&
					strings.ToUpper(artifact.Value.(string)) == hostedZoneId {
					errMsg := "There is a route53 project extension with inputted subdomain already."
					log.InfoWithFields(errMsg, log.Fields{
						"project_extension_id":          projectExtension.Model.ID.String(),
						"existing_project_extension_id": existingProjectExtension.Model.ID.String(),
						"environment_id":                projectExtension.EnvironmentID.String(),
						"hosted_zone_id":                hostedZoneId,
					})
					return fmt.Errorf(errMsg)
				}
			}
		}
	}

	return nil
}
