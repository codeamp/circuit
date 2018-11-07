package codeamp

import (
	"github.com/codeamp/transistor"
	"github.com/codeamp/circuit/plugins"
	"github.com/davecgh/go-spew/spew"
	"github.com/codeamp/circuit/plugins/codeamp/model"
)

func (x *CodeAmp) SmartProfiles(project *model.Project) error {
	spew.Dump("SmartProfiles")

	payload := plugins.Project{}

	x.Events <- transistor.NewEvent(plugins.GetEventName("smartprofiles"), transistor.GetAction("create"), payload)

	// spew.Dump(ev)

	return nil
}