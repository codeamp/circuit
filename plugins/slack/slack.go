package slack

import (
	"fmt"
	"strings"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	slack "github.com/lytics/slackhook"

	log "github.com/codeamp/logger"
)

//Slack is a local struct for slack plugin
type Slack struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("slack", func() transistor.Plugin {
		return &Slack{}
	}, plugins.NotificationExtension{}, plugins.ProjectExtension{})
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
	if e.Name == plugins.GetEventName("slack") && e.Action == transistor.GetAction("status") {
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

	messageStatus, _ := e.GetArtifact("message")
	if err != nil {
		return err
	}
	dashboardURL, err := e.GetArtifact("dashboard_url")
	if err != nil {
		return err
	}

	tail := payload.Release.TailFeature.Hash
	head := payload.Release.HeadFeature.Hash

	showGithubCompareUrl := true
	releaseFeatureHash := fmt.Sprintf("%s ... %s", tail[:8], head[:8])
	if tail == head {
		releaseFeatureHash = head[:8]
		showGithubCompareUrl = false
	}

	var resultColor string
	switch status := strings.ToLower(messageStatus.String()); status {
	case "failed":
		resultColor = "#FF0000"
	case "canceled":
		resultColor = "#9400D3"
	case "success":
		resultColor = "#008000"
	}

	deployMessage := fmt.Sprintf("%s - %s", strings.ToUpper(messageStatus.String()), payload.Environment)
	githubCompareURL := fmt.Sprintf("<https://github.com/%s/compare/%s...%s|%s ... %s>", payload.Project.Repository, tail, head, tail[:8], head[:8])
	githubLinkUrl := fmt.Sprintf("<https://github.com/%s/commit/%s>", payload.Project.Repository, head)
	dashboardPath := fmt.Sprintf("<%s/projects/%s/%s/releases|%s>", dashboardURL.String(), payload.Project.Slug, payload.Environment, payload.Project.Repository)
	releaseMessage := fmt.Sprintf("Project <%s|%s> _%s_", githubLinkUrl, releaseFeatureHash, payload.Release.HeadFeature.Message)

	text := deployMessage + "\n"
	text = text + "\n" + releaseMessage
	if showGithubCompareUrl {
		text = text + "\n" + githubCompareURL
	}
	text = text + "\n" + dashboardPath

	header := fmt.Sprintf("Deployed %s", payload.Project.Repository)
	resultAttachments := slack.Attachment{
		Color:     resultColor,
		Text:      text,
		Footer:    fmt.Sprintf("By: %s", payload.Release.User),
		Title:     fmt.Sprintf("%s - Release", payload.Project.Repository),
		TitleLink: fmt.Sprintf("https://github.com/%s", payload.Project.Repository),
	}

	slackPayload := slack.Message{
		UserName:  "CodeAmp",
		Channel:   fmt.Sprintf("#%s", channel.String()),
		IconEmoji: ":rocket:",
		Text:      header,
	}

	slackPayload.AddAttachment(&resultAttachments)

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

	payload := slack.Message{
		Text:      fmt.Sprintf("Installed slack webhook to %s/%s", ePayload.Environment, ePayload.Project.Repository),
		UserName:  "CodeAmp",
		Channel:   fmt.Sprintf("#%s", channel),
		IconEmoji: fmt.Sprintf(":rocket:"),
	}

	webHookErr := sendSlackMessage(webhook, payload)
	if webHookErr != nil {
		return fmt.Errorf("webhook_url: %s is invalid. Valid webhook_url is required. ErrorMessage: %s", webhook, webHookErr)
	}

	return nil
}

func sendSlackMessage(webhook string, payload slack.Message) error {
	client := slack.New(webhook)
	slackErr := client.Send(&payload)
	if slackErr != nil {
		return fmt.Errorf("Slack Notification failed to dispatch! %s", slackErr)
	}
	return nil
}
