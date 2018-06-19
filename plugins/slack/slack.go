package slack

import (
	"fmt"

	slack "github.com/ashwanthkumar/slack-go-webhook"
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
)

//Slack is a local struct for slack plugin
type Slack struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("slack", func() transistor.Plugin {
		return &Slack{}
	}, plugins.NotificationExtension{})
}

// Description: Plugin description
func (x *Slack) Description() string {
	return "Emit slack events on certain release status"
}

// SampleConfig return plugin sample config
func (x *Slack) SampleConfig() string {
	return ` `
}

// Start plugin
func (x *Slack) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Slack Notifier")
	return nil
}

// Stop spins slack down
func (x *Slack) Stop() {
	log.Info("Stopping Slack Notifier")
}

// Subscribe to events
func (x *Slack) Subscribe() []string {
	return []string{
		"slack:create",
		"slack:update",
		"slack:delete",
		"slack:status",
		"slack:notify",
	}
}

// Process slack webhook events
func (x *Slack) Process(e transistor.Event) error {
	log.DebugWithFields("Processing Slack event", log.Fields{
		"event": e.Event(),
	})

	// no-op on slack plugin statuses
	if e.Name == plugins.GetEventName("slack:status") {
		return nil
	}

	webHookURL, err := e.GetArtifact("webhook_url")
	if err != nil {
		return err
	}

	channel, err := e.GetArtifact("channel")
	if err != nil {
		return err
	}

	if e.Action == transistor.GetAction("create") || e.Action == transistor.GetAction("update") {
		validationErr := validateSlackWebhook(webHookURL.String(), channel.String(), e)
		if validationErr != nil {
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), validationErr.Error())
			return validationErr
		}

		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "")
		return nil
	}

	if e.Name != plugins.GetEventName("slack:notify") {
		return nil
	}

	payload := e.Payload.(plugins.NotificationExtension)

	icon := ""
	iconArtifact, err := e.GetArtifact("emoji")
	if err != nil {
		icon = ":rocket:"
	} else {
		icon = iconArtifact.String()
	}

	messageStatus, _ := e.GetArtifact("message")

	message := fmt.Sprintf("%s deployed %s/%s - Status: %s", payload.Release.User, payload.Environment, payload.Project.Repository, messageStatus.String())

	slackPayload := slack.Payload{
		Text:        message,
		Username:    "CodeAmp",
		Channel:     fmt.Sprintf("#%s", channel.String()),
		IconEmoji:   fmt.Sprintf("%s", icon),
		Attachments: []slack.Attachment{},
	}

	ev := e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Successfully sent message")

	slackErr := sendSlackMessage(webHookURL.String(), slackPayload)
	if slackErr != nil {
		errMsg := fmt.Sprintf("Slack Notification failed to dispatch! %s", slackErr.Error())
		ev.State = transistor.GetState("failed")
		ev.StateMessage = errMsg
		x.events <- ev
		return fmt.Errorf(errMsg)
	}

	ev.State = transistor.GetState("complete")
	ev.StateMessage = fmt.Sprintf("Slack Notification Dispatched")
	x.events <- ev

	return nil
}

func validateSlackWebhook(webhook string, channel string, e transistor.Event) error {
	ePayload := e.Payload.(plugins.ProjectExtension)

	payload := slack.Payload{
		Text:        fmt.Sprintf("Installed slack webhook to %s/%s", ePayload.Environment, ePayload.Project.Repository),
		Username:    "Codeamp",
		Channel:     fmt.Sprintf("#%s", channel),
		IconEmoji:   fmt.Sprintf(":rocket:"),
		Attachments: []slack.Attachment{},
	}

	webHookErr := sendSlackMessage(webhook, payload)

	if webHookErr != nil {
		return fmt.Errorf("webhook_url: %s is invalid. Valid webhook_url is required. ErrorMessage: %s", webhook, webHookErr)
	}

	return nil
}

func sendSlackMessage(webhook string, payload slack.Payload) error {
	slackErr := slack.Send(webhook, "", payload)
	if len(slackErr) > 0 {
		return fmt.Errorf("Slack Notification failed to dispatch! %s", slackErr)
	}
	return nil
}
