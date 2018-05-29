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

func (x *CodeAmp) ReleaseEventHandler(e transistor.Event) error {
	var err error
	payload := e.Payload.(plugins.Release)
	release := resolvers.Release{}
	releaseExtensions := []resolvers.ReleaseExtension{}

	if x.DB.Where("id = ?", payload.ID).First(&release).RecordNotFound() {
		log.InfoWithFields("release not found", log.Fields{
			"id": payload.ID,
		})
		return fmt.Errorf("release %s not found", payload.ID)
	}

	if e.Matches("release:create") {
		x.DB.Where("release_id = ?", release.Model.ID).Find(&releaseExtensions)

		for _, releaseExtension := range releaseExtensions {
			projectExtension := resolvers.ProjectExtension{}
			if x.DB.Where("id = ?", releaseExtension.ProjectExtensionID).Find(&projectExtension).RecordNotFound() {
				log.InfoWithFields("project extensions not found", log.Fields{
					"id": releaseExtension.ProjectExtensionID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("project extension %s not found", releaseExtension.ProjectExtensionID)
			}

			extension := resolvers.Extension{}
			if x.DB.Where("id= ?", projectExtension.ExtensionID).Find(&extension).RecordNotFound() {
				log.InfoWithFields("extension not found", log.Fields{
					"id": projectExtension.Model.ID,
					"release_extension_id": releaseExtension.Model.ID,
				})
				return fmt.Errorf("extension %s not found", projectExtension.ExtensionID)
			}

			if plugins.Type(extension.Type) == plugins.GetType("workflow") {
				// check if the last release extension has the same
				// ServicesSignature and SecretsSignature. If so,
				// mark the action as completed before sending the event
				lastReleaseExtension := resolvers.ReleaseExtension{}
				artifacts := []transistor.Artifact{}

				eventAction := transistor.GetAction("create")
				eventState := transistor.GetState("waiting")
				eventStateMessage := ""

				// check if can cache workflows
				if !release.ForceRebuild && !x.DB.Where("project_extension_id = ? and services_signature = ? and secrets_signature = ? and feature_hash = ? and state in (?)", projectExtension.Model.ID, releaseExtension.ServicesSignature, releaseExtension.SecretsSignature, releaseExtension.FeatureHash, []string{"complete"}).Order("created_at desc").First(&lastReleaseExtension).RecordNotFound() {
					eventAction = transistor.GetAction("status")
					eventState = lastReleaseExtension.State
					eventStateMessage = lastReleaseExtension.StateMessage

					err := json.Unmarshal(lastReleaseExtension.Artifacts.RawMessage, &artifacts)
					if err != nil {
						log.Error(err.Error())
					}
				} else {
					artifacts, err = resolvers.ExtractArtifacts(projectExtension, extension, x.DB)
					if err != nil {
						log.Error(err.Error())
					}
				}

				payload := plugins.ReleaseExtension{
					ID:      releaseExtension.Model.ID.String(),
					Release: payload,
				}

				ev := transistor.NewEvent(transistor.EventName(fmt.Sprintf("release:%s", extension.Key)), eventAction, payload)
				ev.State = eventState
				ev.StateMessage = eventStateMessage
				ev.Artifacts = artifacts

				x.Events <- ev
			}
		}
	}
	return nil
}

func (x *CodeAmp) ReleaseFailed(release *resolvers.Release, stateMessage string) {
	release.State = transistor.GetState("failed")
	release.StateMessage = stateMessage
	x.DB.Save(release)

	releaseExtensions := []resolvers.ReleaseExtension{}
	x.DB.Where("release_id = ? AND state <> ?", release.Model.ID, transistor.GetState("complete")).Find(&releaseExtensions)
	for _, re := range releaseExtensions {
		re.State = transistor.GetState("failed")
		x.DB.Save(&re)
	}

	x.RunQueuedReleases(release)
}

func (x *CodeAmp) ReleaseCompleted(release *resolvers.Release) {
	project := resolvers.Project{}
	environment := resolvers.Environment{}

	if x.DB.Where("id = ?", release.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"release": release,
		})
	}

	if x.DB.Where("id = ?", release.EnvironmentID).First(&environment).RecordNotFound() {
		log.WarnWithFields("Environment not found", log.Fields{
			"id": release.EnvironmentID,
		})
	}

	// mark release as complete
	release.State = transistor.GetState("complete")
	release.StateMessage = "Completed"

	x.DB.Save(release)

	payload := plugins.WebsocketMsg{
		Event:   fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
		Payload: release,
	}
	event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
	event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key), false)
	x.Events <- event

	x.RunQueuedReleases(release)
}

func (x *CodeAmp) RunQueuedReleases(release *resolvers.Release) error {
	var nextQueuedRelease resolvers.Release

	if x.DB.Where("id != ? and state = ? and project_id = ? and environment_id = ? and created_at > ?", release.Model.ID, "waiting", release.ProjectID, release.EnvironmentID, release.CreatedAt).Order("created_at asc").First(&nextQueuedRelease).RecordNotFound() {
		log.WarnWithFields("No queued releases found.", log.Fields{
			"id":             release.Model.ID,
			"state":          "waiting",
			"project_id":     release.ProjectID,
			"environment_id": release.EnvironmentID,
			"created_at":     release.CreatedAt,
		})
		return nil
	}

	var project resolvers.Project
	var services []resolvers.Service
	var secrets []resolvers.Secret

	projectSecrets := []resolvers.Secret{}
	// get all the env vars related to this release and store
	x.DB.Where("environment_id = ? AND project_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID, "project").Find(&projectSecrets)
	for _, secret := range projectSecrets {
		var secretValue resolvers.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	globalSecrets := []resolvers.Secret{}
	x.DB.Where("environment_id = ? AND scope = ?", nextQueuedRelease.EnvironmentID, "global").Find(&globalSecrets)
	for _, secret := range globalSecrets {
		var secretValue resolvers.SecretValue
		x.DB.Where("secret_id = ?", secret.Model.ID).Order("created_at desc").First(&secretValue)
		secret.Value = secretValue
		secrets = append(secrets, secret)
	}

	x.DB.Where("project_id = ? and environment_id = ?", nextQueuedRelease.ProjectID, nextQueuedRelease.EnvironmentID).Find(&services)
	if len(services) == 0 {
		log.WarnWithFields("no services found", log.Fields{
			"project_id": nextQueuedRelease.ProjectID,
		})
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	for i, service := range services {
		ports := []resolvers.ServicePort{}
		x.DB.Where("service_id = ?", service.Model.ID).Find(&ports)
		services[i].Ports = ports
	}

	if x.DB.Where("id = ?", nextQueuedRelease.ProjectID).First(&project).RecordNotFound() {
		log.WarnWithFields("project not found", log.Fields{
			"id": nextQueuedRelease.ProjectID,
		})
		return nil
	}

	// get all branches relevant for the project
	var branch string
	var projectSettings resolvers.ProjectSettings

	if x.DB.Where("environment_id = ? and project_id = ?", nextQueuedRelease.EnvironmentID, nextQueuedRelease.ProjectID).First(&projectSettings).RecordNotFound() {
		log.WarnWithFields("no env project branch found", log.Fields{})
	} else {
		branch = projectSettings.GitBranch
	}

	var environment resolvers.Environment
	if x.DB.Where("id = ?", nextQueuedRelease.EnvironmentID).Find(&environment).RecordNotFound() {
		log.WarnWithFields("no env found", log.Fields{
			"id": nextQueuedRelease.EnvironmentID,
		})
		return nil
	}

	var headFeature resolvers.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.HeadFeatureID).First(&headFeature).RecordNotFound() {
		log.WarnWithFields("head feature not found", log.Fields{
			"id": nextQueuedRelease.HeadFeatureID,
		})
		return nil
	}

	var tailFeature resolvers.Feature
	if x.DB.Where("id = ?", nextQueuedRelease.TailFeatureID).First(&tailFeature).RecordNotFound() {
		log.WarnWithFields("tail feature not found", log.Fields{
			"id": nextQueuedRelease.TailFeatureID,
		})
		return nil
	}

	var pluginServices []plugins.Service
	for _, service := range services {
		var spec resolvers.ServiceSpec
		if x.DB.Where("id = ?", service.ServiceSpecID).First(&spec).RecordNotFound() {
			log.WarnWithFields("servicespec not found", log.Fields{
				"id": service.ServiceSpecID,
			})
			return nil
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

	releaseEvent := plugins.Release{
		ID:          nextQueuedRelease.Model.ID.String(),
		Environment: environment.Key,
		HeadFeature: plugins.Feature{
			ID:         headFeature.Model.ID.String(),
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
		User: nextQueuedRelease.User.Email,
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
		Services: pluginServices, // ADB Added this
	}

	x.Events <- transistor.NewEvent(plugins.GetEventName("release"), transistor.GetAction("create"), releaseEvent)
	return nil
}
