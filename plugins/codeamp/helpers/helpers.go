package helpers

import (
	"encoding/json"

	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
)

/* fills in Config by querying config ids and getting the actual value */
func ExtractArtifacts(projectExtension model.ProjectExtension, extension model.Extension, db *gorm.DB) ([]transistor.Artifact, error) {
	var artifacts []transistor.Artifact
	var err error

	extensionConfig := []model.ExtConfig{}
	if extension.Config.RawMessage != nil {
		err = json.Unmarshal(extension.Config.RawMessage, &extensionConfig)
		if err != nil {
			return nil, err
		}
	}

	projectConfig := []model.ExtConfig{}
	if projectExtension.Config.RawMessage != nil {
		err = json.Unmarshal(projectExtension.Config.RawMessage, &projectConfig)
		if err != nil {
			return nil, err
		}
	}

	existingArtifacts := []transistor.Artifact{}
	if projectExtension.Artifacts.RawMessage != nil {
		err = json.Unmarshal(projectExtension.Artifacts.RawMessage, &existingArtifacts)
		if err != nil {
			return nil, err
		}
	}

	for i, ec := range extensionConfig {
		for _, pc := range projectConfig {
			if ec.AllowOverride && ec.Key == pc.Key && pc.Value != "" {
				extensionConfig[i].Value = pc.Value
			}
		}

		var artifact transistor.Artifact
		// check if val is UUID. If so, query in environment variables for id
		secretID := uuid.FromStringOrNil(extensionConfig[i].Value)
		if secretID != uuid.Nil {
			secret := model.SecretValue{}
			if db.Where("secret_id = ?", secretID).Order("created_at desc").First(&secret).RecordNotFound() {
				log.InfoWithFields("secret not found", log.Fields{
					"secret_id": secretID,
				})
			}
			artifact.Key = ec.Key
			artifact.Value = secret.Value
		} else {
			artifact.Key = ec.Key
			artifact.Value = extensionConfig[i].Value
		}
		artifacts = append(artifacts, artifact)
	}

	for _, ea := range existingArtifacts {
		artifacts = append(artifacts, ea)
	}

	projectCustomConfig := make(map[string]interface{})
	if projectExtension.CustomConfig.RawMessage != nil {
		err = json.Unmarshal(projectExtension.CustomConfig.RawMessage, &projectCustomConfig)
		if err != nil {
			log.Error(err.Error())
			return nil, err
		}
	}

	for key, val := range projectCustomConfig {
		var artifact transistor.Artifact
		artifact.Key = key
		artifact.Value = val
		artifact.Secret = false
		artifacts = append(artifacts, artifact)
	}

	return artifacts, nil
}
