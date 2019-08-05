package codeamp

import (
	"context"
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"

	"github.com/jinzhu/gorm"
)

func (x *CodeAmp) GitSync(project *model.Project) error {
	var feature model.Feature
	var release model.Release
	var headFeature model.Feature
	hash := ""

	// Get latest release and deployed feature hash
	if x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&feature)
		hash = feature.Hash
	} else {
		if x.DB.Where("id = ?", release.HeadFeatureID).Find(&headFeature).RecordNotFound() {
			log.InfoWithFields("can not find head feature", log.Fields{
				"id": release.HeadFeatureID,
			})
		}
		hash = headFeature.Hash
	}

	// get branches of entire environments
	projectSettingsCollection := []model.ProjectSettings{}
	if x.DB.Where("project_id = ?", project.Model.ID.String()).Find(&projectSettingsCollection).RecordNotFound() {
		payload := plugins.GitSync{
			Project: plugins.Project{
				ID:         project.Model.ID.String(),
				Repository: project.Repository,
			},
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        "master",
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
			From: hash,
		}

		x.Events <- transistor.NewEvent(plugins.GetEventName("gitsync"), transistor.GetAction("create"), payload)
	} else {
		for _, projectSettings := range projectSettingsCollection {
			payload := plugins.GitSync{
				Project: plugins.Project{
					ID:         project.Model.ID.String(),
					Repository: project.Repository,
				},
				Git: plugins.Git{
					Url:           project.GitUrl,
					Protocol:      project.GitProtocol,
					Branch:        projectSettings.GitBranch,
					RsaPrivateKey: project.RsaPrivateKey,
					RsaPublicKey:  project.RsaPublicKey,
				},
				From: hash,
			}

			x.Events <- transistor.NewEvent(plugins.GetEventName("gitsync"), transistor.GetAction("create"), payload)
		}
	}

	return nil
}

func (x *CodeAmp) GitSyncEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.GitSync)

	var project model.Project
	var projectSettings []model.ProjectSettings
	var features []model.Feature

	if e.State == transistor.GetState("complete") {
		if x.DB.Where("repository = ?", payload.Project.Repository).First(&project).RecordNotFound() {
			log.ErrorWithFields("project not found", log.Fields{
				"repository": payload.Project.Repository,
			})
			return nil
		}

		releaseNotificationWebsocketEvents := make([]string, 0, 10)

		foundFeatures := len(payload.Commits)
		newFeatures := 0
		for _, commit := range payload.Commits {
			var feature model.Feature
			if x.DB.Where("project_id = ? AND hash = ?", project.ID, commit.Hash).First(&feature).RecordNotFound() {
				feature = model.Feature{
					ProjectID:  project.ID,
					Message:    commit.Message,
					User:       commit.User,
					Hash:       commit.Hash,
					ParentHash: commit.ParentHash,
					Ref:        commit.Ref,
					Created:    commit.Created,
				}

				if err := x.DB.Save(&feature).Error; err != nil {
					log.Error(err.Error())
				}

				newFeatures = newFeatures + 1

				if commit.Head {
					if err := x.DB.Where("project_id = ?", project.Model.ID).Find(&projectSettings).Error; err != nil {
						if gorm.IsRecordNotFoundError(err) {
							log.ErrorWithFields("No project settings found", log.Fields{
								"project_id": project.Model.ID,
							})
						} else {
							log.Error(err.Error())
						}
					} else {
						// Create an automated release if specified by the projects configuration/settings
						// call CreateRelease for each env that has cd turned on
						for _, setting := range projectSettings {
							var environment model.Environment
							if x.DB.Where("id = ?", setting.EnvironmentID).First(&environment).RecordNotFound() {
								log.ErrorWithFields("Environment not found", log.Fields{
									"id": setting.EnvironmentID,
								})
								continue
							}

							// Notify listeners there are new features found, but only for envs with this branch set
							if newFeatures > 0 && fmt.Sprintf("refs/heads/%s", setting.GitBranch) == feature.Ref {
								releaseNotificationWebsocketEvents = append(releaseNotificationWebsocketEvents, fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key))
							}

							if setting.ContinuousDeploy == true {
								if setting.ContinuousDeploy && fmt.Sprintf("refs/heads/%s", setting.GitBranch) == feature.Ref {
									adminContext := context.WithValue(context.Background(), "jwt", model.Claims{
										UserID:      uuid.FromStringOrNil(ContinuousDeployUUID).String(),
										Email:       "codeamp@codeamp.com",
										Permissions: []string{"admin"},
									})

									x.Resolver.CreateRelease(adminContext, &struct {
										Release *model.ReleaseInput
									}{
										Release: &model.ReleaseInput{
											HeadFeatureID: feature.Model.ID.String(),
											ProjectID:     setting.ProjectID.String(),
											EnvironmentID: setting.EnvironmentID.String(),
										},
									})

									releaseNotificationWebsocketEvents = append(releaseNotificationWebsocketEvents, fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key))
									releaseNotificationWebsocketEvents = append(releaseNotificationWebsocketEvents, fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key))
								}
							}
						}
					}
				}
			}
		}

		// put payload commits' hash in a map for quicker query below
		payloadCommits := make(map[string]bool)
		for _, commit := range payload.Commits {
			payloadCommits[commit.Hash] = true
		}

		// limit the features array to 50, becaues payload.Commits is also of max length 50
		if x.DB.Where("project_id = ? AND ref = ? AND not_found_since IS NULL", project.ID, fmt.Sprintf("refs/heads/%s", payload.Git.Branch)).Order("created desc").Limit(50).Find(&features).RecordNotFound() {
			log.ErrorWithFields("Feature not found", log.Fields{
				"project_id": project.ID,
			})
			return nil
		}

		// handle forced push commits
		for _, feature := range features {
			if _, ok := payloadCommits[feature.Hash]; ok {
				continue
			}

			// set NotFoundSince for the unfound feature
			if feature.NotFoundSince == nil {
				notFoundSince := time.Now()
				feature.NotFoundSince = &notFoundSince
				if err := x.DB.Save(&feature).Error; err != nil {
					log.Error(err.Error())
				}
			}

			// found a related release
			var release model.Release
			if !x.DB.Where("project_id = ? AND head_feature_id = ?", project.ID, feature.ID).First(&release).RecordNotFound() {
				release.Redeployable = false
				release.RedeployableMessage = "This feature cannot be found."
				if err := x.DB.Save(&release).Error; err != nil {
					log.Error(err.Error())
				}
			}
		}

		log.Debug(fmt.Sprintf("Sync: [%s] - Found %d features. %d were new.", project.GitUrl, foundFeatures, newFeatures))
		for _, msg := range releaseNotificationWebsocketEvents {
			payload := plugins.WebsocketMsg{
				Event: msg,
			}

			event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
			event.AddArtifact("event", msg, false)
			x.Events <- event
		}
	}

	return nil
}
