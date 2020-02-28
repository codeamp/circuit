package codeamp

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	uuid "github.com/satori/go.uuid"

	"github.com/jinzhu/gorm"
)

func (x *CodeAmp) GitSync(project *model.Project) error {
	var feature model.Feature
	var release model.Release
	var headFeature model.Feature
	hash := ""

	// Get latest release and deployed feature hash
	log.Debug("Grabbing a gitsync")
	if x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.DB.Where("project_id = ?", project.ID).Order("created_at DESC").First(&feature)

		log.Warn("there was no release found for the project: ", project.ID)
		spew.Dump(feature)

		hash = feature.Hash
	} else {
		log.Warn("Looking for head feature: ", release.HeadFeatureID)
		if x.DB.Where("id = ?", release.HeadFeatureID).Find(&headFeature).RecordNotFound() {
			log.InfoWithFields("can not find head feature", log.Fields{
				"id": release.HeadFeatureID,
			})
		}
		hash = headFeature.Hash
	}

	// get branches of entire environments
	// projectSettingsCollection := []model.ProjectSettings{}
	projectSettings := model.ProjectSettings{}
	environmentsList := []model.Environment{}

	log.Warn("Finding environments list")
	if err := x.DB.Find(&environmentsList).Error; err != nil {
		log.Error(err.Error())
	}
	spew.Dump(environmentsList)

	hasProjectSettings := false
	for _, environment := range environmentsList {
		log.Warn("looping through env: ", environment.Key)
		if err := x.DB.Where("project_id = ? AND environment_id = ?", project.Model.ID.String(), environment.ID.String()).
			Order("created_at").First(&projectSettings).Error; err != nil {

			if gorm.IsRecordNotFoundError(err) == false {
				log.Error(err.Error())
			}
			spew.Dump(err)

		} else {
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

			hasProjectSettings = true
		}
	}

	if hasProjectSettings == false {
		log.Warn("PROJECT HAS NO PROJECT SETTINGS ASSIGNED! - ", project.Name)
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
	}

	return nil
}

func (x *CodeAmp) GitSyncEventHandler(e transistor.Event) error {
	payload := e.Payload.(plugins.GitSync)

	var project model.Project
	var projectSettings []model.ProjectSettings

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
			x.DB.LogMode(true)
			defer x.DB.LogMode(false)
			if x.DB.Where("project_id = ? AND hash = ? AND ref = ?", project.ID, commit.Hash, commit.Ref).First(&feature).RecordNotFound() {
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
				} else {
					log.Warn("Saving feature for ", commit.Ref)
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
						spew.Dump(projectSettings)

						log.Warn("checking for automated release status")
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

			x.DB.LogMode(false)
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
