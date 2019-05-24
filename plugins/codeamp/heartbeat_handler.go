package codeamp

import (
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

func (x *CodeAmp) HeartBeatEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.HeartBeat)

	var projects []model.Project

	if err := x.DB.Find(&projects).Error; err != nil {
		log.Error(err.Error())
	}
	for _, project := range projects {
		switch payload.Tick {
		case "minute":
			x.GitSync(&project)
		}
	}

	switch payload.Tick {
	case "minute":
		err := x.scheduledBranchReleaserPulse(e)
		if err != nil {
			log.Error(err.Error())
		}
	}
}
