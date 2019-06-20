package codeamp

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	"github.com/jinzhu/gorm/dialects/postgres"
)

func (x *CodeAmp) ReleaseExtensionEventHandler(e transistor.Event) error {
	log.Warn("ReleaseExtensionEventHandler")
	payload := e.Payload.(plugins.ReleaseExtension)

	var releaseExtension model.ReleaseExtension
	var release model.Release

	if e.Matches("release:.*:status") {
		if x.DB.Where("id = ?", payload.Release.ID).Find(&release).RecordNotFound() {
			log.InfoWithFields("release", log.Fields{
				"id": payload.Release.ID,
			})
			return fmt.Errorf("Release %s not found", payload.Release.ID)
		}

		if x.DB.Where("id = ?", payload.ID).Find(&releaseExtension).RecordNotFound() {
			log.InfoWithFields("release extension not found", log.Fields{
				"id": payload.ID,
			})
			return fmt.Errorf("Release extension %s not found", payload.ID)
		}

		releaseExtension.State = e.State
		releaseExtension.StateMessage = e.StateMessage
		marshalledReArtifacts, err := json.Marshal(e.Artifacts)
		if err != nil {
			log.Error(err.Error())
			return err
		}

		releaseExtension.Artifacts = postgres.Jsonb{marshalledReArtifacts}
		releaseExtension.Finished = time.Now()
		if err := x.DB.Save(&releaseExtension).Error; err != nil {
			log.Error(err)
			return err
		}

		switch e.State {
		case transistor.GetState("complete"):
			x.ReleaseExtensionCompleted(&payload, &releaseExtension)
		case transistor.GetState("failed"):
			x.ReleaseFailed(&release, e.StateMessage)
		case transistor.GetState("running"):
			break
		default:
			log.Error("RELEASE IN ERROR STATE: ", e.State)
		}
	}

	return nil
}

func (x *CodeAmp) ReleaseExtensionCompleted(payload *plugins.ReleaseExtension, re *model.ReleaseExtension) {
	log.Warn("ReleaseExtensionCompleted")

	releaseExtensions := []model.ReleaseExtension{}
	if x.DB.Where("release_id = ?", re.ReleaseID).Find(&releaseExtensions).RecordNotFound() {
		log.ErrorWithFields("release extensions not found", log.Fields{
			"releaseExtension": re,
		})
		return
	}

	// Notify clientside a release extension has completed
	{
		wssPayload := plugins.WebsocketMsg{
			Event: fmt.Sprintf("projects/%s/%s/releases/reCompleted", payload.Release.Project.Slug, payload.Release.Environment),
		}
		event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), wssPayload)
		event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases/reCompleted", payload.Release.Project.Slug, payload.Release.Environment), false)

		x.Events <- event
	}

	// loop through and check if all same-type release extensions are completed
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == re.Type && releaseExtension.State != transistor.GetState("complete") {
			return
		}
	}

	var release model.Release
	if x.DB.Where("id = ?", payload.Release.ID).First(&release).RecordNotFound() {
		log.WarnWithFields("release not found", log.Fields{
			"release": release,
		})
		return
	}

	switch re.Type {
	case plugins.GetType("workflow"):
		x.WorkflowReleaseExtensionsCompleted(payload, &release)
	case plugins.GetType("deployment"):
		log.Warn("Releases completed")
		x.ReleaseCompleted(&release)
	}
}

func (x *CodeAmp) WorkflowReleaseExtensionsCompleted(payload *plugins.ReleaseExtension, release *model.Release) {
	project := model.Project{}
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": release.ProjectID,
		})
		return
	}

	headFeature := model.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := model.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := model.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	user := model.User{}
	email := ""

	specialReleaseUsers := map[string]string{
		ContinuousDeployUUID: "Automated Deployment",
		ScheduledDeployUUID:  "Scheduled Deployment",
	}
	if val, ok := specialReleaseUsers[release.UserID.String()]; ok {
		email = val
	} else {
		if err := x.DB.Where("id = ?", release.UserID).First(&user).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
				log.InfoWithFields("user not found", log.Fields{
					"id": release.UserID,
				})
			} else {
				log.Error(err.Error())
			}
		} else {
			email = user.Email
		}
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := model.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.WarnWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var secrets []model.Secret
	err := json.Unmarshal(release.Secrets.RawMessage, &secrets)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		return
	}

	var services []model.Service
	err = json.Unmarshal(release.Services.RawMessage, &services)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		return
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec model.ServiceSpec
		if x.DB.Where("service_id = ?", service.Model.ID).First(&spec).RecordNotFound() {
			log.WarnWithFields("servicespec not found", log.Fields{
				"service_id": service.Model.ID,
			})
			return
		}

		pluginServices = graphql_resolver.AppendPluginService(pluginServices, service, spec)
	}

	var pluginSecrets []plugins.Secret
	for _, secret := range secrets {
		pluginSecrets = append(pluginSecrets, plugins.Secret{
			Key:   secret.Key,
			Value: secret.Value.Value,
			Type:  secret.Type,
		})
	}

	// insert CodeAmp envs
	slugSecret := plugins.Secret{
		Key:   "CODEAMP_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, slugSecret)

	hashSecret := plugins.Secret{
		Key:   "CODEAMP_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, hashSecret)

	timeSecret := plugins.Secret{
		Key:   "CODEAMP_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, timeSecret)

	// insert Codeflow envs - remove later
	_slugSecret := plugins.Secret{
		Key:   "CODEFLOW_SLUG",
		Value: project.Slug,
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _slugSecret)

	_hashSecret := plugins.Secret{
		Key:   "CODEFLOW_HASH",
		Value: headFeature.Hash[0:7],
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _hashSecret)

	_timeSecret := plugins.Secret{
		Key:   "CODEFLOW_CREATED_AT",
		Value: time.Now().Format(time.RFC3339),
		Type:  plugins.GetType("env"),
	}
	pluginSecrets = append(pluginSecrets, _timeSecret)

	releaseExtensionDeploymentsCount := 0
	/**************************************
	*
	* COLLECT WORKFLOW ARTIFACTS
	*
	/*************************************/
	releaseExtensions := []model.ReleaseExtension{}
	artifacts := []transistor.Artifact{}

	if err := x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions).Error; err != nil {
		log.Error(err)
	}
	for _, releaseExtension := range releaseExtensions {
		if releaseExtension.Type == plugins.GetType("workflow") {
			projectExtension := model.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id":                   releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := model.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id":                   projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			var unmarshalledArtifacts []transistor.Artifact
			err := json.Unmarshal(releaseExtension.Artifacts.RawMessage, &unmarshalledArtifacts)
			if err != nil {
				log.ErrorWithFields(err.Error(), log.Fields{})
				return
			}

			for _, artifact := range unmarshalledArtifacts {
				artifact.Source = extension.Key
				artifacts = append(artifacts, artifact)
			}
		}
	}

	/**************************************
	*
	* SEND 'DEPLOYMENT' TYPE RE EVENTS
	*
	**************************************/
	for _, releaseExtension := range releaseExtensions {
		releaseExtensionAction := transistor.GetAction("create")
		if releaseExtension.Type == plugins.GetType("deployment") {
			_artifacts := artifacts

			// Fail deployment if the release is in a failed state
			if release.State == transistor.GetState("failed") {
				releaseExtensionAction = transistor.GetAction("status")
				releaseExtension.State = transistor.GetState("failed")
				releaseExtension.StateMessage = release.StateMessage

				if err := x.DB.Save(&releaseExtension); err != nil {
					log.Error(err)
				}
				log.Error("RELEASE IN FAILED STATE")
				continue
			}

			projectExtension := model.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id":                   releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := model.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id":                   projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			projectExtensionArtifacts, err := graphql_resolver.ExtractArtifacts(projectExtension, extension, x.DB)
			if err != nil {
				log.Error(err.Error())
			}

			for _, artifact := range projectExtensionArtifacts {
				_artifacts = append(_artifacts, artifact)
			}

			releaseExtension.Started = time.Now()
			if err := x.DB.Save(&releaseExtension).Error; err != nil {
				log.Error(err)
			}

			releaseExtensionEvent := plugins.ReleaseExtension{
				ID: releaseExtension.Model.ID.String(),
				Release: plugins.Release{
					ID:          release.Model.ID.String(),
					Environment: environment.Key,
					HeadFeature: plugins.Feature{
						Hash:       headFeature.Hash,
						ParentHash: headFeature.ParentHash,
						User:       headFeature.User,
						Message:    headFeature.Message,
						Created:    headFeature.Created,
					},
					TailFeature: plugins.Feature{
						ID:         tailFeature.Model.ID.String(),
						Hash:       tailFeature.Hash,
						ParentHash: tailFeature.ParentHash,
						User:       tailFeature.User,
						Message:    tailFeature.Message,
						Created:    tailFeature.Created,
					},
					User: email,
					Project: plugins.Project{
						ID:         project.Model.ID.String(),
						Slug:       project.Slug,
						Repository: project.Repository,
					},
					Git: plugins.Git{
						Url:           project.GitUrl,
						Branch:        branch,
						RsaPrivateKey: project.RsaPrivateKey,
					},
					Secrets:    pluginSecrets,
					Services:   pluginServices,
					IsRollback: release.IsRollback,
				},
			}

			ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), releaseExtensionAction, releaseExtensionEvent)
			ev.Artifacts = _artifacts
			x.Events <- ev

			releaseExtensionDeploymentsCount++
		}
	}

	log.Warn("releaseExtensionDeploymentsCount ", releaseExtensionDeploymentsCount)
	if releaseExtensionDeploymentsCount == 0 {
		x.ReleaseCompleted(release)
	}
}
