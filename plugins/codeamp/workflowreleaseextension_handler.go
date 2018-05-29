package codeamp

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/codeamp/circuit/plugins"
	resolvers "github.com/codeamp/circuit/plugins/codeamp/resolvers"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

func (x *CodeAmp) WorkflowReleaseExtensionsCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"id": release.ProjectID,
		})
		return
	}

	headFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.InfoWithFields("head feature not found", log.Fields{
			"id": release.HeadFeatureID,
		})
		return
	}

	tailFeature := resolvers.Feature{}
	if x.DB.Where("id = ?", release.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.InfoWithFields("tail feature not found", log.Fields{
			"id": release.TailFeatureID,
		})
		return
	}

	environment := resolvers.Environment{}
	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.InfoWithFields("environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
		return
	}

	user := resolvers.User{}
	if x.DB.Where("id = ?", release.UserID).First(&user).RecordNotFound() {
		log.InfoWithFields("user not found", log.Fields{
			"id": release.UserID,
		})
	}

	// get all branches relevant for the projec
	branch := "master"
	projectSettings := resolvers.ProjectSettings{}
	if x.DB.Where("environment_id = ? and project_id = ?", environment.Model.ID.String(),
		project.Model.ID.String()).First(&projectSettings).RecordNotFound() {
		log.WarnWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var secrets []resolvers.Secret
	err := json.Unmarshal(release.Secrets.RawMessage, &secrets)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		return
	}

	var services []resolvers.Service
	err = json.Unmarshal(release.Services.RawMessage, &services)
	if err != nil {
		log.Error(err.Error(), log.Fields{})
		return
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec resolvers.ServiceSpec
		if x.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.WarnWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return
		}

		listeners := []plugins.Listener{}
		for _, l := range service.Ports {
			p, err := strconv.ParseInt(l.Port, 10, 32)
			if err != nil {
				panic(err)
			}
			listener := plugins.Listener{
				Port:     int32(p),
				Protocol: l.Protocol,
			}
			listeners = append(listeners, listener)
		}

		pluginServices = resolvers.AppendPluginService(pluginServices, service, listeners, spec)
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
	releaseExtensions := []resolvers.ReleaseExtension{}
	artifacts := []transistor.Artifact{}

	x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)
	for _, releaseExtension := range releaseExtensions {
		// collect workflow artifacts
		if releaseExtension.Type == plugins.GetType("workflow") {
			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
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

	for _, releaseExtension := range releaseExtensions {
		releaseExtensionAction := transistor.GetAction("create")
		if releaseExtension.Type == plugins.GetType("deployment") {
			_artifacts := artifacts

			// Fail deployment if the release is in a failed state
			if release.State == transistor.GetState("failed") {
				releaseExtensionAction = transistor.GetAction("status")
				releaseExtension.State = transistor.GetState("failed")
				releaseExtension.StateMessage = release.StateMessage
			}

			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.WarnWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.WarnWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return
			}

			projectExtensionArtifacts, err := resolvers.ExtractArtifacts(projectExtension, extension, x.DB)
			if err != nil {
				log.Error(err.Error())
			}

			for _, artifact := range projectExtensionArtifacts {
				_artifacts = append(_artifacts, artifact)
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
					User: user.Email,
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
					Secrets:  pluginSecrets,
					Services: pluginServices,
				},
			}

			ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), releaseExtensionAction, releaseExtensionEvent)
			ev.Artifacts = _artifacts
			x.Events <- ev

			releaseExtensionDeploymentsCount++
		}
	}

	if releaseExtensionDeploymentsCount == 0 {
		x.ReleaseCompleted(release)
	}
}
