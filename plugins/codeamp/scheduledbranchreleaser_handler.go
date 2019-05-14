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

func (x *CodeAmp) ScheduledBranchReleaserEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.ScheduledBranchReleaser)
	log.Warn("ScheduledBranchReleaserEventHandler - ", e.Event())

	desiredBranch, err := e.GetArtifact("branch")
	if err != nil {
		log.Error(err.Error())
		return
	}

	// Find the project settings for the project in question
	var projectSettings model.ProjectSettings
	if err := x.DB.Where("id = ?", payload.ProjectSettingsID).Find(&projectSettings).Error; err != nil {
		log.Error(err.Error())
	} else {
		if projectSettings.GitBranch != desiredBranch.String() {
			var environment model.Environment
			if err := x.DB.Where("id = ?", payload.Environment).Find(&environment).Error; err != nil {
				log.Error(err.Error())
				return
			}

			// Switch to desired branch
			{
				log.Warn("WARNING: UPDATING GITBRANCH TO DESIRED BRANCH: ", projectSettings.GitBranch, " ", desiredBranch.String())

				projectSettings.GitBranch = desiredBranch.String()
				if err := x.DB.Save(&projectSettings).Error; err != nil {
					log.Error(err.Error())
				}
			}

			// Pull in commits for this branch
			headFeatureID := ""
			{
				_, err := x.commits(payload.ProjectExtension.Project.Repository, payload.Git)
				if err != nil {
					log.Error(err)
					return
				}

				var release model.Release
				var feature model.Feature

				if err := x.DB.Where("project_id = ? AND environment_id = ?", projectSettings.ProjectID, projectSettings.EnvironmentID).Order("created_at DESC").First(&release).Error; err != nil {
					if gorm.IsRecordNotFoundError(err) {
						if err := x.DB.Where("project_id = ? AND ref = ", projectSettings.ProjectID, fmt.Sprintf("refs/heads/%s", projectSettings.GitBranch)).Order("created_at DESC").First(&feature).Error; err != nil {
							log.Error(err.Error())
							return
						}
						headFeatureID = feature.Model.ID.String()
					} else {
						log.Error(err.Error())
						return
					}
				} else {
					if err := x.DB.Where("id = ?", release.HeadFeatureID).Find(&feature).Error; err != nil {
						if gorm.IsRecordNotFoundError(err) {
							log.InfoWithFields("can not find head feature", log.Fields{
								"id": release.HeadFeatureID,
							})
						} else {
							log.Error(err.Error())
							return
						}
					}
					headFeatureID = release.HeadFeatureID.String()
				}
			}

			// If this succeeds, send a message to the front end to make user refresh the 'settings' page if they are viewing it
			{
				payload := plugins.WebsocketMsg{
					Event: fmt.Sprintf("project/branch-update"),
				}

				log.Warn("Sending Message: ", "project/branch-update")
				event := transistor.NewEvent(plugins.GetEventName("websocket"), transistor.GetAction("status"), payload)
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
			}
		}
	}
}

func (x *CodeAmp) commits(projectRepository string, git plugins.Git) ([]plugins.GitCommit, error) {
	var err error
	var output []byte

	idRsaPath := fmt.Sprintf("%s/%s_id_rsa", viper.GetString("plugins.gitsync.workdir"), projectRepository)
	idRsa := fmt.Sprintf("GIT_SSH_COMMAND=ssh -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no -i %s -F /dev/null", idRsaPath)
	repoPath := fmt.Sprintf("%s/%s_%s", viper.GetString("plugins.gitsync.workdir"), projectRepository, git.Branch)

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
		commit, err := x.toGitCommit(line, head)
		if err != nil {
			log.Error(err)
			return nil, err
		}

		commits = append(commits, commit)
	}

	return commits, nil
}

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

func (x *CodeAmp) toGitCommit(entry string, head bool) (plugins.GitCommit, error) {
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
	}, nil
}

func (x *CodeAmp) scheduledBranchReleaserPulse(e transistor.Event) error {
	type Data struct {
		Branch        string
		EnvironmentID string
		ExtensionID   string
	}

	data := Data{"master", "85d85e59-5260-42fb-ab6d-d5d5aafbfe4f", "f45d29ca-059e-4112-8b65-1001f06f9425"}

	var extension model.Extension
	if err := x.DB.Where("id = ?", data.ExtensionID).Find(&extension).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	var projectExtensions []model.ProjectExtension
	if err := x.DB.Where("environment_id = ? AND extension_id = ?", data.EnvironmentID, data.ExtensionID).Find(&projectExtensions).Error; err != nil {
		log.Error(err.Error())
		return err
	}

	var projectSettings model.ProjectSettings
	var project model.Project
	for _, peResult := range projectExtensions {
		if err := x.DB.Where("git_branch != ? AND environment_id = ? AND project_id = ?", data.Branch, data.EnvironmentID, peResult.ProjectID).Find(&projectSettings).Error; err != nil {
			if gorm.IsRecordNotFoundError(err) == false {
				log.Error(err.Error())
			}

			continue
		}

		err := x.DB.Where("id = ?", projectSettings.ProjectID).Find(&project).Error
		if err != nil {
			log.Error(err.Error())
			continue
		}

		artifacts, err := graphql_resolver.ExtractArtifacts(peResult, extension, x.DB)
		if err != nil {
			log.Error(err.Error())
			continue
		}

		projectSchedulerExtension := plugins.ProjectExtension{
			ID: peResult.Model.ID.String(),
			Project: plugins.Project{
				ID:         projectSettings.ProjectID.String(),
				Slug:       project.Slug,
				Repository: project.Repository,
			},
			Environment: data.EnvironmentID,
		}

		log.Warn("SENDING PULSE MESSAGE")
		sbr := plugins.ScheduledBranchReleaser{
			ProjectSettingsID: projectSettings.Model.ID.String(),
			ProjectExtension:  projectSchedulerExtension,
			Git: plugins.Git{
				Url:           project.GitUrl,
				Protocol:      project.GitProtocol,
				Branch:        data.Branch,
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
