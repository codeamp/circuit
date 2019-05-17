package codeamp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/codeamp/circuit/plugins"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
	uuid "github.com/satori/go.uuid"
	"github.com/spf13/viper"
)

// This has to be here because we need a way of referencing the extension
// in order to get the extension ID and then grab all the project extensions
// that have that extension id as their parent extension
const (
	SCHEDULED_RELEASE_HANDLER_PLUGIN_NAME = "scheduledbranchreleaser"
)

// This is the handler that is automagically called when
// CodeAmp receives a message with a 'ScheduleBranchReleaser' type payload
//
// It's job is to handle the message that is being received from the plugin itself.
// If this function is executing, then the plugin has determined that the current
// time matches the scheduled time for this project.
//
// The job of this function is to build a release with the desired branch
// and to call the graphql function that is responsible for building a release
func (x *CodeAmp) ScheduledBranchReleaserEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.ScheduledBranchReleaser)
	desiredBranch, err := e.GetArtifact("branch")
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Find the project settings for the project in question
	var projectSettings model.ProjectSettings
	if err := x.DB.Where("id = ?", payload.ProjectSettingsID).Find(&projectSettings).Error; err != nil {
		log.Error(err.Error())
		return
	} else {
		if projectSettings.GitBranch != desiredBranch.String() {
			oldBranch := projectSettings.GitBranch

			// Try and find the envrionment that is associted with the environment id that was provided
			// in the message. this is necessary so we can send a message to the front end to inform
			// the user that the branch has been updated without their explicit input
			var environment model.Environment
			if err := x.DB.Where("id = ?", payload.Environment).Find(&environment).Error; err != nil {
				log.Error(err.Error())
				return
			}

			// Update the project settings DB entry to reflect the new branch selection
			{
				projectSettings.GitBranch = desiredBranch.String()
				if err := x.DB.Save(&projectSettings).Error; err != nil {
					log.Error(err.Error())
					return
				} else {
					log.WarnWithFields("[AUDIT] Updated Project Branch (Automated)", log.Fields{
						"project":     payload.Project.Slug,
						"branch":      desiredBranch.String(),
						"oldBranch":   oldBranch,
						"user":        "scheduled builder",
						"environment": payload.Environment},
					)
				}
			}

			// Pull in all commits for this branch. Don't rely on gitsync for this operation
			// as that would require adding SBR specific events into the Gitsync plugin
			// since only 1 receiver will handle any given message (no multi receives)
			headFeatureID := ""
			{
				commits, err := x.commits(payload.ProjectExtension.Project.Repository, payload.Git)
				if err != nil {
					log.Error(err.Error())
					return
				}

				err = x.loadCommits(&payload, commits, environment.Name)
				if err != nil {
					log.Error(err.Error())
					return
				}

				var release model.Release
				var feature model.Feature

				// Try to grab the latest release in order to determine the most recently deployed feature hash
				// if that doesn't exist, then try and find the most recent commit for the branch in the list of features
				if err := x.DB.Where("project_id = ? AND environment_id = ?", projectSettings.ProjectID, projectSettings.EnvironmentID).Order("created_at DESC").First(&release).Error; err != nil {
					if gorm.IsRecordNotFoundError(err) {
						if err := x.DB.Where("project_id = ? AND ref = ?", projectSettings.ProjectID, fmt.Sprintf("refs/heads/%s", projectSettings.GitBranch)).Order("created_at DESC").First(&feature).Error; err != nil {
							log.Error(err.Error())
							return
						}
						headFeatureID = feature.Model.ID.String()
					} else {
						log.Error(err.Error())
						return
					}
				} else {
					headFeatureID = release.HeadFeatureID.String()
				}
			}

			// If this succeeds, send a message to the front end to make user refresh the 'settings' page if they are viewing it
			{
				event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), plugins.WebsocketMsg{
					Event: "project/branch-update",
				})
				event.AddArtifact("event", "msg", false)
				x.Events <- event
			}

			// Proceed to trigger a build from the graphql interface
			{
				adminContext := context.WithValue(context.Background(), "jwt", model.Claims{
					UserID:      uuid.FromStringOrNil(ScheduledDeployUUID).String(),
					Email:       "codeamp@codeamp.com",
					Permissions: []string{"admin"},
				})

				releaseInput := &model.ReleaseInput{
					HeadFeatureID: headFeatureID,
					ProjectID:     payload.ProjectExtension.Project.ID,
					EnvironmentID: payload.Environment,
					ForceRebuild:  false,
				}
				_, err := x.Resolver.CreateRelease(adminContext, &struct{ Release *model.ReleaseInput }{releaseInput})
				if err != nil {
					log.Error(err.Error())
					return
				}

				event := transistor.NewEvent(plugins.GetEventName("scheduledbranchreleaser:scheduled"), transistor.GetAction("status"), payload)
				event.State = transistor.GetState("complete")
				event.StateMessage = "ScheduledBranchReleaser Scheduled Release"
				x.Events <- event
			}
		}
	}
}

func (x *CodeAmp) loadCommits(payload *plugins.ScheduledBranchReleaser, commits []plugins.GitCommit, environmentName string) error {
	var project model.Project
	if x.DB.Where("repository = ?", payload.Project.Repository).First(&project).RecordNotFound() {
		log.ErrorWithFields("project not found", log.Fields{
			"repository": payload.Project.Repository,
		})
		return nil
	}

	newFeatures := 0
	for _, commit := range commits {
		var feature model.Feature
		if err := x.DB.Where("project_id = ? AND hash = ?", project.ID, commit.Hash).Find(&feature).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) {
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
					return err
				}
				newFeatures += 1
			} else {
				log.Error(err.Error())
			}
		}
	}

	// Notify listeners there are new features found, but only for envs with this branch set
	if newFeatures > 0 {
		payload := plugins.WebsocketMsg{Event: fmt.Sprintf("projects/%s/%s/features", payload.Project.Slug, environmentName)}
		event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)

		x.Events <- event
	}

	return nil
}

// Pulled from the Gitsync plugin
func (x *CodeAmp) commits(projectRepository string, git plugins.Git) ([]plugins.GitCommit, error) {
	var err error
	var output []byte

	idRsaPath := fmt.Sprintf("%s/%s_id_rsa", viper.GetString("plugins.scheduledbranchreleaser.workdir"), projectRepository)
	idRsa := fmt.Sprintf("GIT_SSH_COMMAND=ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i %s -F /dev/null", idRsaPath)
	repoPath := fmt.Sprintf("%s/%s_%s", viper.GetString("plugins.scheduledbranchreleaser.workdir"), projectRepository, git.Branch)

	// Git Env
	env := os.Environ()
	env = append(env, idRsa)

	_, err = exec.Command("mkdir", "-p", filepath.Dir(repoPath)).CombinedOutput()
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(idRsaPath); os.IsNotExist(err) {
		log.InfoWithFields("creating repository id_rsa", log.Fields{
			"path": idRsaPath,
		})

		err := ioutil.WriteFile(idRsaPath, []byte(git.RsaPrivateKey), 0600)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	if _, err := os.Stat(repoPath); os.IsNotExist(err) {
		log.InfoWithFields("cloning repository", log.Fields{
			"path": repoPath,
		})

		_, err := x.git(env, "clone", git.Url, repoPath)
		if err != nil {
			log.Error(err)
			return nil, err
		}
	}

	output, err = x.git(env, "-C", repoPath, "reset", "--hard", fmt.Sprintf("origin/%s", git.Branch))
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "clean", "-fd")
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "pull", "origin", git.Branch)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "checkout", git.Branch)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	output, err = x.git(env, "-C", repoPath, "log", "--first-parent", "--date=iso-strict", "-n", "50", "--pretty=format:%H#@#%P#@#%s#@#%cN#@#%cd", git.Branch)

	if err != nil {
		log.Error(err.Error())
		return nil, err
	}

	var commits []plugins.GitCommit
	for i, line := range strings.Split(strings.TrimSuffix(string(output), "\n"), "\n") {
		head := false
		if i == 0 {
			head = true
		}
		commit, err := x.toGitCommit(line, head, git.Branch)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

// Pulled from the Gitsync plugin
func (x *CodeAmp) git(env []string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)

	cmd.Env = env
	out, err := cmd.CombinedOutput()
	if err != nil {
		if ee, ok := err.(*exec.Error); ok {
			if ee.Err == exec.ErrNotFound {
				return nil, errors.New("Git executable not found in $PATH")
			}
		}

		return nil, errors.New(string(bytes.TrimSpace(out)))
	}

	return out, nil
}

// Pulled from the Gitsync plugin
func (x *CodeAmp) toGitCommit(entry string, head bool, branch string) (plugins.GitCommit, error) {
	items := strings.Split(entry, "#@#")
	commiterDate, err := time.Parse("2006-01-02T15:04:05-07:00", items[4])

	if err != nil {
		return plugins.GitCommit{}, err
	}

	return plugins.GitCommit{
		Hash:       items[0],
		ParentHash: items[1],
		Message:    items[2],
		User:       items[3],
		Head:       head,
		Created:    commiterDate,
		Ref:        fmt.Sprintf("refs/heads/%s", branch),
	}, nil
}

// This function is called on receipt of a hearbeat message from the
// hearbeat plugin. The purpose of this function is to gather
// all configurations of all project extensions which have the SBR
// as the base extension. From there, the function checks to see if the
// desired branch is different from the current branch. If so, it sends a pulse
// message to the SBR plugin, where we then will check the schedule to see if it's
// the appropriate time to update the branch and create a release
func (x *CodeAmp) scheduledBranchReleaserPulse(e transistor.Event) error {
	// Unfortunately we have to rely on the name of the extension here to grab the base extension
	// this is CRUCIAL and if the plugin is added to the codeamp extensions list then this handler
	// WILL NOT FUNCTION. Seriously.
	var extension model.Extension
	if err := x.DB.Where("key = ?", SCHEDULED_RELEASE_HANDLER_PLUGIN_NAME).Find(&extension).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	// Use the extension id to find all project extensions that have been setup
	var projectExtensions []model.ProjectExtension
	if err := x.DB.Where("extension_id = ?", extension.ID.String()).Find(&projectExtensions).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	// Iterate over all the found project extensions and build the necessary configuration
	// We'll need to pull the configuration from the project extension and the base extension
	// and merge them together. From there we'll need to find the project settings
	// where the environment and project id matches, but it is not currently set to our
	// desired branch. In this case, send a pulse message to the plugin in order for the plugin
	// to determine if it is the appropriate time to create a release
	var projectSettings model.ProjectSettings
	var project model.Project

	for _, peResult := range projectExtensions {
		err := x.DB.Where("id = ?", peResult.ProjectID).Find(&project).Error
		if err != nil {
			log.Error(err.Error())
			continue
		}

		// Extract and merge the project extension, and extension artifacts
		artifacts, err := graphql_resolver.ExtractArtifacts(peResult, extension, x.DB)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		desiredBranch := ""
		for _, artifact := range artifacts {
			if artifact.Key == "BRANCH" {
				desiredBranch = artifact.String()
			}
		}

		if desiredBranch == "" {
			log.Error(errors.New("Coud not find desired branch in plugin extension configuration"))
			continue
		}

		// We want to grab the project settings where the project and environment match what we're looking for
		// but the currently configured branch is different than our desired branch we want to trigger a build on
		if err := x.DB.Where("git_branch != ? AND environment_id = ? AND project_id = ?", desiredBranch, peResult.EnvironmentID, peResult.ProjectID).Find(&projectSettings).Error; err != nil {
			// It's okay if this comes back without finding any records.
			// only report an error if there is an actual error other than no records found
			if gorm.IsRecordNotFoundError(err) == false {
				log.Error(err.Error())
			}

			// We need to continue regardless of the cause of the error condition
			// because we don't have enough information to build out the ReleaseInput struct
			continue
		}

		// We'll need to build a project extension payload,
		// as well as a Gitsync payload and the artifacts we've extracted
		// to send off to the SBR plugin
		projectSchedulerExtension := plugins.ProjectExtension{
			ID: peResult.Model.ID.String(),
			Project: plugins.Project{
				ID:         projectSettings.ProjectID.String(),
				Slug:       project.Slug,
				Repository: project.Repository,
			},
			Environment: peResult.EnvironmentID.String(),
		}

		sbr := plugins.ScheduledBranchReleaser{
			ProjectSettingsID: projectSettings.Model.ID.String(),
			ProjectExtension:  projectSchedulerExtension,
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        desiredBranch,
				RsaPrivateKey: project.RsaPrivateKey,
				RsaPublicKey:  project.RsaPublicKey,
			},
		}

		event := transistor.NewEvent(plugins.GetEventName("scheduledbranchreleaser:pulse"), transistor.GetAction("status"), sbr)
		event.Artifacts = artifacts

		x.Events <- event
	}

	return nil
}
