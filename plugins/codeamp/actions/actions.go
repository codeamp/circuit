package codeamp_actions

import (
	"fmt"

	"github.com/codeamp/circuit/plugins"
	codeamp_models "github.com/codeamp/circuit/plugins/codeamp/models"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/jinzhu/gorm"
)

type Actions struct {
	events chan transistor.Event
	db     *gorm.DB
}

func NewActions(events chan transistor.Event, db *gorm.DB) *Actions {
	return &Actions{
		events: events,
		db:     db,
	}
}

func (x *Actions) HeartBeat(tick string) {
	var projects []codeamp_models.Project

	x.db.Find(&projects)
	for _, project := range projects {
		if tick == "minute" {
			x.GitSync(&project)
		}
	}
}

func (x *Actions) GitSync(project *codeamp_models.Project) {
	var feature codeamp_models.Feature
	var release codeamp_models.Release
	hash := ""

	// Get latest release and deployed feature hash
	if x.db.Where("project_id = ?", project.ID).Order("created_at DESC").First(&release).RecordNotFound() {
		// get latest feature if there is no releases
		x.db.Where("project_id = ?", project.ID).Order("created DESC").First(&feature)
		hash = feature.Hash
	} else {
		hash = release.HeadFeature.Hash
	}

	gitSync := plugins.GitSync{
		Action: plugins.Update,
		State:  plugins.Waiting,
		Project: plugins.Project{
			Slug:       project.Slug,
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

	x.events <- transistor.NewEvent(gitSync, nil)
}

func (x *Actions) GitCommit(commit plugins.GitCommit) {
	project := codeamp_models.Project{}
	feature := codeamp_models.Feature{}

	if x.db.Where("repository = ?", commit.Repository).First(&project).RecordNotFound() {
		log.InfoWithFields("project not found", log.Fields{
			"repository": commit.Repository,
		})
		return
	}

	if x.db.Where("project_id = ? AND hash = ?", project.ID, commit.Hash).First(&feature).RecordNotFound() {
		feature = codeamp_models.Feature{
			ProjectId:  project.ID,
			Message:    commit.Message,
			User:       commit.User,
			Hash:       commit.Hash,
			ParentHash: commit.ParentHash,
			Ref:        commit.Ref,
			Created:    commit.Created,
		}
		x.db.Save(&feature)

		wsMsg := plugins.WebsocketMsg{
			Event:   fmt.Sprintf("projects/%s/features", project.Slug),
			Payload: feature,
		}
		x.events <- transistor.NewEvent(wsMsg, nil)
	} else {
		log.InfoWithFields("feature already exists", log.Fields{
			"repository": commit.Repository,
			"hash":       commit.Hash,
		})
	}
}

func (x *Actions) ProjectCreated(project *codeamp_models.Project) {
	wsMsg := plugins.WebsocketMsg{
		Event:   "projects",
		Payload: project,
	}
	x.events <- transistor.NewEvent(wsMsg, nil)
}
