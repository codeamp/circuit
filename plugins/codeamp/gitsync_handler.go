package codeamp

import (
	"context"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
)

func (x *CodeAmp) GitSync(project *graphql_resolver.Project) error {
	var feature graphql_resolver.Feature
	var release model.Release
	var headFeature graphql_resolver.Feature
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
	projectSettingsCollection := []graphql_resolver.ProjectSettings{}
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

	var project graphql_resolver.Project
	var environment graphql_resolver.Environment
	var projectSettings []graphql_resolver.ProjectSettings

	if e.State == transistor.GetState("complete") {
		if x.DB.Where("repository = ?", payload.Project.Repository).First(&project).RecordNotFound() {
			log.ErrorWithFields("project not found", log.Fields{
				"repository": payload.Project.Repository,
			})
			return nil
		}

		for _, commit := range payload.Commits {
			var feature graphql_resolver.Feature
			if x.DB.Where("project_id = ? AND hash = ?", project.ID, commit.Hash).First(&feature).RecordNotFound() {
				feature = graphql_resolver.Feature{
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
									Email:       "codeamp",
									Permissions: []string{"admin"},
								})

								x.Resolver.CreateRelease(adminContext, &struct {
									Release *graphql_resolver.ReleaseInput
								}{
									Release: &graphql_resolver.ReleaseInput{
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

							payload := plugins.WebsocketMsg{
								Event: fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key),
							}
							event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
							event.AddArtifact("event", fmt.Sprintf("projects/%s/%s/features", project.Slug, environment.Key), false)
							x.Events <- event
						}
					}
				}
			} else {
				log.WarnWithFields("feature already exists", log.Fields{
					"repository": commit.Repository,
					"hash":       commit.Hash,
				})
			}
		}
	}

	return nil
}
