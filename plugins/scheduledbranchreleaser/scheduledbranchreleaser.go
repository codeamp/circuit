package scheduledbranchreleaser

import (
	"fmt"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

const (
	PULSE_MESSAGE   = "scheduledbranchreleaser:pulse"
	RELEASE_MESSAGE = "scheduledbranchreleaser:release"

	SCHEDULED_TIME_THRESHOLD = time.Duration(time.Minute * 45)
)

func init() {
	transistor.RegisterPlugin("scheduledbranchreleaser", func() transistor.Plugin {
		return &ScheduledBranchReleaser{}
	}, plugins.ProjectExtension{}, plugins.ScheduledBranchReleaser{})
}

func (x *ScheduledBranchReleaser) Description() string {
	return "Switch branch back to master and deploy project"
}

func (x *ScheduledBranchReleaser) SampleConfig() string {
	return ` `
}

func (x *ScheduledBranchReleaser) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started ScheduledBranchReleaser")
	return nil
}

func (x *ScheduledBranchReleaser) Stop() {
	log.Info("Stopping ScheduledBranchReleaser")
}

func (x *ScheduledBranchReleaser) Subscribe() []string {
	return []string{
		"project:scheduledbranchreleaser:create",
		"project:scheduledbranchreleaser:update",
		"project:scheduledbranchreleaser:delete",
		"scheduledbranchreleaser:pulse",
	}
}

// How does this work?
// This plugin mostly relies on the CodeAmp plugin in order
// to make the magic happen.
//
// This plugin receives a 'pulse' message from the CodeAmp plugin
// after that plugin receive a 'heartbeat' message. The pulse
// message contains data for the project extension. This plugin
// uses that information to determine if we should proceed through the function
// based solely on if the schedule matches up with the desired build time.
//
// The CodeAmp plugin is responsible for dispatching these messages
// and for determining if the project is in an applicable configuration
// that being a branch being set to something other than the desired BRANCH
//
// Once this plugin has decided that the schedule matches, it fires off an event
// that is then again handled by the CodeAmp plugin. The second event handling
// is the portion that is responsible for orchetrasting the release
// and triggering the release process.
//
// Required Fields:
// BRANCH		- Which branch should this project be automatically updated to?
// SCHEDULE		- When should the branch be updated and a release created?
func (x *ScheduledBranchReleaser) Process(e transistor.Event) error {
	var err error
	if e.Matches("project:scheduledbranchreleaser") {
		log.InfoWithFields(fmt.Sprintf("Process ScheduledBranchReleaser event: %s", e.Event()), log.Fields{})
		switch e.Action {
		case transistor.GetAction("create"):
			// err = x.createScheduledBranchReleaser(e)
		case transistor.GetAction("update"):
			// err = x.updateScheduledBranchReleaser(e)
		case transistor.GetAction("delete"):
			// err = x.deleteScheduledBranchReleaser(e)
		default:
			log.Warn(fmt.Sprintf("Unhandled ScheduledBranchReleaser event: %s", e.Event()))

		}

		x.sendResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), "Nothing to Update. Removing this extension does not delete any data.", nil)

		if err != nil {
			log.Error(err.Error())
			return err
		}
	} else if e.Matches(PULSE_MESSAGE) {
		payload := e.Payload.(plugins.ScheduledBranchReleaser)
		timeScheduledToBuild, err := e.GetArtifact("schedule")
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// TODO: Remove hardcoded time
		log.Warn(timeScheduledToBuild.String())
		t, err := time.Parse("15:04 -0700 MST", "22:00 -0700 UTC")
		if err != nil {
			log.Error(err.Error())
			return err
		}

		now := time.Now().UTC()
		parsedDuration, err := time.ParseDuration(fmt.Sprintf("%dh%dm", t.Hour(), t.Minute()))
		if err != nil {
			log.Error(err.Error())
			return err
		}
		scheduledTime := now.Truncate(time.Hour * 24).Add(parsedDuration)

		nowDiff := scheduledTime.Sub(now)

		if nowDiff <= SCHEDULED_TIME_THRESHOLD {
			event := transistor.NewEvent(RELEASE_MESSAGE, transistor.GetAction("create"), payload)
			event.Artifacts = e.Artifacts
			x.events <- event
		}
	}

	return nil
}

// Wrapper for sending an event back thruogh the messaging system for standardization and brevity
func (x *ScheduledBranchReleaser) sendResponse(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, artifacts []transistor.Artifact) {
	event := e.NewEvent(action, state, stateMessage)
	event.Artifacts = artifacts

	x.events <- event
}
