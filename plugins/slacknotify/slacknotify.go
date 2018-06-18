package slacknotify

import (
	"fmt"
	"io/ioutil"
	"net/http"

	slack "github.com/ashwanthkumar/slack-go-webhook"
	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"

	log "github.com/codeamp/logger"
)

//Slack is a local struct for slack plugin
type Slack struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("slacknotify", func() transistor.Plugin {
		return &Slack{}
	}, plugins.ProjectExtension{})
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
		"project:slack:create",
		"project:slack:update",
		"project:slack:delete",
		"slacknotify",
	}
}

// Process slack webhook events
func (x *Slack) Process(e transistor.Event) error {
	log.DebugWithFields("Processing Slack event", log.Fields{
		"event": e.Event(),
	})

	webHookURL, err := e.GetArtifact("webhook_url")
	if err != nil {
		return err
	}

	spew.Dump(webHookURL)

	payload := e.Payload.(plugins.ProjectExtension)

	spew.Dump(payload)

	if e.Action == transistor.GetAction("create") || e.Action == transistor.GetAction("update") {
		// return nil
		// projectExtension := plugins.ProjectExtension{
		// 	Environment: lbPayload.Environment,
		// 	Project:     lbPayload.Project,
		// }

		event := e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), "Plugin failed")

		projectExtension := plugins.ProjectExtension{
			Environment: payload.Environment,
			Project:     payload.Project,
			ID:          payload.ID,
		}
		spew.Dump(payload)
		event.SetPayload(projectExtension)
		x.events <- event

		// err = validateSlackWebhook(webHookURL.String())
		// if err != nil {
		// 	x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), "")
		// 	return err
		// }

		// x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "")
	}

	return nil

	if e.State != transistor.GetState("complete") && e.State != transistor.GetState("failed") {
		return nil
	}

	channel := ""
	channelArtifact, err := e.GetArtifact("channel")
	if err != nil {
		return fmt.Errorf("channel is not set. Valid slack channel is required")
	}
	channel = channelArtifact.String()

	icon := ""
	iconArtifact, err := e.GetArtifact("emoji")
	if err != nil {
		icon = ":rocket:"
	} else {
		icon = iconArtifact.String()
	}

	message := ""
	if e.State == transistor.GetState("complete") {
		message = "Deployment for project <PROJECT LINK> succeeded"
	}

	if e.State == transistor.GetState("failed") {
		message = "Deployment for project <PROJECT_LINK> failed"
	}

	slackPayload := slack.Payload{
		Text:        message,
		Username:    "CodeAmp",
		Channel:     fmt.Sprintf("#%s", channel),
		IconEmoji:   fmt.Sprintf("%s", icon),
		Attachments: []slack.Attachment{},
	}

	ev := transistor.NewEvent(plugins.GetEventName("slack"), transistor.GetAction("status"), nil)

	slackErr := slack.Send(webHookURL.String(), "", slackPayload)
	if len(slackErr) > 0 {
		ev.State = transistor.GetState("failed")
		ev.StateMessage = fmt.Sprintf("%s", slackErr)
		x.events <- ev
		return fmt.Errorf("Slack Notification failed to dispatch! %s", slackErr)
	}

	ev.State = transistor.GetState("complete")
	ev.StateMessage = fmt.Sprintf("Slack Notification Dispatched")

	x.events <- ev

	return nil
}

func validateSlackWebhook(webhook string) error {
	var webHookErr string
	webHookErr = ""
	resp, err := http.Get(webhook)
	if err != nil {
		return err
	}
	if resp.StatusCode == 404 {
		defer resp.Body.Close()
		content, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		webHookErr = fmt.Sprintf("Invalid URL - %s", content)

	}

	if webHookErr != "" {
		return fmt.Errorf("webhook_url: %s is invalid. Valid webhook_url is required. Status:%s ErrorMessage: %s", webhook, resp.Status, webHookErr)
	}

	return nil
}
