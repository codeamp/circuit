package codeamp

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
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
	var environment model.Environment
	var projectSettings []model.ProjectSettings

	if e.State == transistor.GetState("complete") {
		if x.DB.Where("repository = ?", payload.Project.Repository).First(&project).RecordNotFound() {
			log.ErrorWithFields("project not found", log.Fields{
				"repository": payload.Project.Repository,
			})
			return nil
		}

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

				x.DB.Save(&feature)

				if commit.Head {
					if x.DB.Where("continuous_deploy = ? and project_id = ?", true, project.Model.ID).Find(&projectSettings).RecordNotFound() {
						log.ErrorWithFields("No continuous deploys found", log.Fields{
							"continuous_deploy": true,
							"project_id":        project.Model.ID,
						})
					} else {
						// call CreateRelease for each env that has cd turned on
						for _, setting := range projectSettings {
							if setting.ContinuousDeploy && fmt.Sprintf("refs/heads/%s", setting.GitBranch) == feature.Ref {
								adminContext := context.WithValue(context.Background(), "jwt", model.Claims{
									UserID:      uuid.FromStringOrNil("codeamp").String(),
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
							}

							if x.DB.Where("id = ?", setting.EnvironmentID).First(&environment).RecordNotFound() {
								log.ErrorWithFields("Environment not found", log.Fields{
									"id": setting.EnvironmentID,
								})
							}
							websocketMsgs := []plugins.WebsocketMsg{
								plugins.WebsocketMsg{
									Event: fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key),
								},
								plugins.WebsocketMsg{
									Event: fmt.Sprintf("projects/%s/%s/releases", project.Slug, environment.Key),
								},
							}
							for _, msg := range websocketMsgs {
								event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
								event.AddArtifact("event", msg.Event, false)
								x.Events <- event
							}
						}
					}
				}
			} else {
				log.DebugWithFields("feature already exists", log.Fields{
					"repository": commit.Repository,
					"hash":       commit.Hash,
				})
			}
		}
	}

	return nil
}
