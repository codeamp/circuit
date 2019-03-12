package helpers

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func CreateProjectExtensionInDB(tx *gorm.DB, input *model.ProjectExtensionInput) (*model.ProjectExtension, error) {
	var projectExtension model.ProjectExtension

	// Check if project can create project extension in environment
	if err := tx.Where("environment_id = ? and project_id = ?", input.EnvironmentID, input.ProjectID).Find(&model.ProjectEnvironment{}).Error; err != nil {
		return nil, fmt.Errorf("Project not allowed to install extensions in given environment")
	}

	extension := model.Extension{}
	if err := tx.Where("id = ?", input.ExtensionID).Find(&extension).Error; err != nil {
		return nil, fmt.Errorf("No extension found for id: '%s'", input.ExtensionID)
	}

	project := model.Project{}
	if err := tx.Where("id = ?", input.ProjectID).Find(&project).Error; err != nil {
		return nil, fmt.Errorf("No project found: '%s'", input.ProjectID)
	}

	env := model.Environment{}
	if err := tx.Where("id = ?", input.EnvironmentID).Find(&env).Error; err != nil {
		return nil, fmt.Errorf("No environment found: '%s'", input.ProjectID)
	}

	// check if extension already exists with project
	// ignore if the extension type is 'once' (installable many times)

	err := tx.Where("project_id = ? and extension_id = ? and environment_id = ?", input.ProjectID, input.ExtensionID, input.EnvironmentID).Find(&projectExtension).Error
	if !gorm.IsRecordNotFoundError(err) {
		log.ErrorWithFields(err.Error(), log.Fields{
			"project_id":     input.ProjectID,
			"extension_id":   input.ExtensionID,
			"environment_id": input.EnvironmentID,
		})
	}

	if extension.Type == plugins.GetType("once") || extension.Type == plugins.GetType("notification") || gorm.IsRecordNotFoundError(err) {
		if extension.Key == "route53" {
			err := HandleExtensionRoute53(tx, input, &projectExtension)
			if err != nil {
				return nil, err
			}
		}

		projectExtension = model.ProjectExtension{
			State:         transistor.GetState("waiting"),
			ExtensionID:   extension.Model.ID,
			ProjectID:     project.Model.ID,
			EnvironmentID: env.Model.ID,
			Config:        postgres.Jsonb{[]byte(input.Config.RawMessage)},
			CustomConfig:  postgres.Jsonb{[]byte(input.CustomConfig.RawMessage)},
		}

		if err := tx.Create(&projectExtension).Error; err != nil {
			return nil, err
		}

		return &projectExtension, nil
	}

	return nil, fmt.Errorf("This extension is already installed in this project.")
}

func HandleExtensionRoute53(tx *gorm.DB, input *model.ProjectExtensionInput, projectExtension *model.ProjectExtension) error {
	extension := model.Extension{}
	if err := tx.Where("id = ?", input.ExtensionID).Find(&extension).Error; err != nil {
		if !gorm.IsRecordNotFoundError(err) {
			log.ErrorWithFields(err.Error(), log.Fields{
				"id": input.ExtensionID,
			})
		}

		return fmt.Errorf("No extension found.")
	}

	// HOTFIX: check for existing subdomains for route53
	unmarshaledCustomConfig := make(map[string]interface{})
	err := json.Unmarshal(input.CustomConfig.RawMessage, &unmarshaledCustomConfig)
	if err != nil {
		return fmt.Errorf("Could not unmarshal custom config")
	}

	artifacts, err := ExtractArtifacts(*projectExtension, extension, tx)
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

	if hostedZoneId == "" {
		return fmt.Errorf("No HOSTED_ZONE_ID Provided")
	}

	existingProjectExtensions := GetProjectExtensionsWithRoute53Subdomain(strings.ToUpper(unmarshaledCustomConfig["subdomain"].(string)), tx)
	for _, existingProjectExtension := range existingProjectExtensions {
		if existingProjectExtension.Model.ID.String() != "" {
			// check if HOSTED_ZONE_ID is the same
			var tmpExtension model.Extension

			if err := tx.Where("id = ?", existingProjectExtension.ExtensionID).First(&tmpExtension).Error; err != nil {
				return err
			}

			tmpExtensionArtifacts, err := ExtractArtifacts(existingProjectExtension, tmpExtension, tx)
			if err != nil {
				return err
			}

			for _, artifact := range tmpExtensionArtifacts {
				if artifact.Key == "HOSTED_ZONE_ID" &&
					strings.ToUpper(artifact.Value.(string)) == hostedZoneId {
					errMsg := "There is a route53 project extension with inputted subdomain already."
					return fmt.Errorf(errMsg)
				}
			}
		}
	}

	return nil
}

func GetProjectExtensionsWithRoute53Subdomain(subdomain string, db *gorm.DB) []model.ProjectExtension {
	var existingProjectExtensions []model.ProjectExtension

	if db.Where("custom_config ->> 'subdomain' ilike ?", subdomain).Find(&existingProjectExtensions).RecordNotFound() {
		return []model.ProjectExtension{}
	}

	return existingProjectExtensions
}
