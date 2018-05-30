package codeamp

import (
	"github.com/codeamp/circuit/plugins"
	resolver "github.com/codeamp/circuit/plugins/graphql/resolver"
	"github.com/codeamp/transistor"
)

func (x *CodeAmp) HeartBeatEventHandler(e transistor.Event) {
	payload := e.Payload.(plugins.HeartBeat)

	var projects []resolver.Project

	x.DB.Find(&projects)
	for _, project := range projects {
		switch payload.Tick {
		case "minute":
			x.GitSync(&project)
		}
	}
}
