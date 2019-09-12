package scheduledbranchreleaser

import (
	"fmt"
	"math"
	"time"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
)

const (
	PULSE_MESSAGE    = "scheduledbranchreleaser:pulse"
	RELEASE_MESSAGE  = "scheduledbranchreleaser:release"
	COMPLETE_MESSAGE = "scheduledbranchreleaser:scheduled"

	SCHEDULED_TIME_THRESHOLD = time.Duration(time.Minute * 1)
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
	x.Events = e
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
		PULSE_MESSAGE,
		COMPLETE_MESSAGE,
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
// and triggering the release process
//
// Required Fields:
// BRANCH		- Which branch should this project be automatically updated to?
// SCHEDULE		- When should the branch be updated and a release created?
func (x *ScheduledBranchReleaser) Process(e transistor.Event) error {
	var err error

	// This message group is sent when a project extension
	// is added or removed from panel
	if e.Matches("project:scheduledbranchreleaser") {
		log.InfoWithFields(fmt.Sprintf("Process ScheduledBranchReleaser event: %s", e.Event()), log.Fields{})

		message := "Unhandled Error"
		switch e.Action {
		case transistor.GetAction("create"):
			message = "Successfully installed extension"
		case transistor.GetAction("update"):
			message = "Successfully updated extension"
		case transistor.GetAction("delete"):
			message = "Deleted Extension"
		default:
			log.Warn(fmt.Sprintf("Unhandled ScheduledBranchReleaser event: %s", e.Event()))
		}

		x.sendResponse(e, transistor.GetAction("status"), transistor.GetState("complete"), message, nil)

		if err != nil {
			log.Error(err.Error())
			return err
		}
		// The pulse message is sent as a result of CodeAmp receiving a heartbeat message
		// and forwarding it to SBR including the configuration of a project extension
		// meeting our criteria (not the desired branch on the env the extension is configured for)
	} else if e.Matches(PULSE_MESSAGE) {
		log.InfoWithFields(fmt.Sprintf("Process ScheduledBranchReleaser event: %s", e.Event()), log.Fields{})

		payload := e.Payload.(plugins.ScheduledBranchReleaser)
		timeScheduledToBuild, err := e.GetArtifact("schedule")
		if err != nil {
			log.Error(err.Error())
			return err
		}

		t, err := time.Parse("15:04 -0700 MST", timeScheduledToBuild.String())
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// Grab the current time and date, then truncate the time so we're left with only today's date
		// Then add the time that we have from the SCHEDULE so we can calculate the duration between
		// now and the scheduled time for today
		now := time.Now().UTC()
		utcT := t.UTC()
		parsedDuration, err := time.ParseDuration(fmt.Sprintf("%dh%dm", utcT.Hour(), utcT.Minute()))
		if err != nil {
			log.Error(err.Error())
			return err
		}

		// Chop off the hour component (by truncate) of today's date
		// then re-add the parsed duration as hours to get to 'todays' time interval
		scheduledTime := now.Truncate(time.Hour * 24).Add(parsedDuration)
		nowDiff := scheduledTime.Sub(now)

		// If the difference between the scheduled time and the now time
		// is less than our time threshold, then send a message back to the CodeAmp plugin
		// in order to create a release for this project
		if math.Abs(nowDiff.Minutes()) <= SCHEDULED_TIME_THRESHOLD.Minutes() {
			event := transistor.NewEvent(RELEASE_MESSAGE, transistor.GetAction("create"), payload)
			event.Artifacts = e.Artifacts
			x.Events <- event
		}
	} else if e.Matches(COMPLETE_MESSAGE) {
		log.Debug("Received complete message: ", e.Event())
	}

	return nil
}

// Wrapper for sending an event back thruogh the messaging system for standardization and brevity
func (x *ScheduledBranchReleaser) sendResponse(e transistor.Event, action transistor.Action, state transistor.State, stateMessage string, artifacts []transistor.Artifact) {
	event := e.NewEvent(action, state, stateMessage)
	event.Artifacts = artifacts

	x.Events <- event
}
