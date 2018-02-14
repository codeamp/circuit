package githubstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"

	log "github.com/codeamp/logger"
)

type GithubStatus struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("githubstatus", func() transistor.Plugin {
		return &GithubStatus{}
	})
}

func (x *GithubStatus) Description() string {
	return "Get status of commit/ build in Github, whether it's LGTM or CircleCi"
}

func (x *GithubStatus) SampleConfig() string {
	return ` `
}

func (x *GithubStatus) Start(e chan transistor.Event) error {
	x.events = e
	log.Println("Started GithubStatus")
	return nil
}

func (x *GithubStatus) Stop() {
	log.Println("Stopping GithubStatus")
}

func (x *GithubStatus) Subscribe() []string {
	return []string{
		"plugins.Extension:create:githubstatus",
		"plugins.ReleaseExtension:create:githubstatus",
	}
}

func (x *GithubStatus) Process(e transistor.Event) error {
	log.InfoWithFields("Processing githubstatus event", log.Fields{
		"event": e,
	})

	if e.Matches("plugins.Extension") {
		event := e.Payload.(plugins.Extension)
		switch event.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus E event: %s", e.Name), log.Fields{})
			responseEvent := e.Payload.(plugins.Extension)
			responseEvent.State = plugins.GetState("complete")
			responseEvent.Action = plugins.GetAction("status")
			x.events <- e.NewEvent(responseEvent, nil)
		}
	}

	if e.Matches("plugins.ReleaseExtension") {
		event := e.Payload.(plugins.ReleaseExtension)
		switch event.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus RE event: %s", e.Name), log.Fields{})
			// get status and check if complete
			client := &http.Client{}
			TIMEOUT_LIMIT := 10
			timeout := 0

			for {
				req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", event.Release.Project.Repository, event.Release.HeadFeature.Hash), nil)
				if err != nil {
					log.Info(err.Error())
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = err.Error()
					x.events <- e.NewEvent(failedEvent, err)
					return nil
				}
				req.SetBasicAuth(event.Extension.Config["GITHUBSTATUS_USERNAME"].(string), event.Extension.Config["GITHUBSTATUS_PERSONAL_ACCESS_TOKEN"].(string))

				resp, _ := client.Do(req)
				if resp.StatusCode == 200 {
					// send an event that we're successfully getting data from github status API
					statusEvent := e.Payload.(plugins.ReleaseExtension)
					statusEvent.State = plugins.GetState("waiting")
					statusEvent.Action = plugins.GetAction("status")
					statusEvent.StateMessage = "Successfully got build events. Waiting for builds to succeed."
					x.events <- e.NewEvent(statusEvent, nil)

					combinedStatusBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						failedEvent := e.Payload.(plugins.ReleaseExtension)
						failedEvent.State = plugins.GetState("failed")
						failedEvent.Action = plugins.GetAction("status")
						failedEvent.StateMessage = err.Error()
						x.events <- e.NewEvent(failedEvent, err)
						return nil
					}
					resp.Body.Close()
					// unmarshal into interface
					var statusBodyInterface interface{}
					if err := json.Unmarshal([]byte(combinedStatusBody), &statusBodyInterface); err != nil {
						failedEvent := e.Payload.(plugins.ReleaseExtension)
						failedEvent.State = plugins.GetState("failed")
						failedEvent.Action = plugins.GetAction("status")
						failedEvent.StateMessage = err.Error()
						x.events <- e.NewEvent(failedEvent, err)
						return nil
					} else {
						if len(statusBodyInterface.(map[string]interface{})["statuses"].([]interface{})) == 0 || statusBodyInterface.(map[string]interface{})["state"].(string) == "success" {
							responseEvent := e.Payload.(plugins.ReleaseExtension)
							responseEvent.State = plugins.GetState("complete")
							responseEvent.Action = plugins.GetAction("status")
							responseEvent.StateMessage = "Completed"
							x.events <- e.NewEvent(responseEvent, nil)
							return nil
						}

						if statusBodyInterface.(map[string]interface{})["state"].(string) == "failed" {
							failedBuilds := ""
							// check which builds failed
							for _, build := range statusBodyInterface.(map[string]interface{})["statuses"].([]interface{}) {
								if build.(map[string]interface{})["state"].(string) == "failed" {
									failedBuilds += fmt.Sprintf(", %s", build.(map[string]interface{})["context"].(string))
								}
							}
							responseEvent := e.Payload.(plugins.ReleaseExtension)
							responseEvent.State = plugins.GetState("failed")
							responseEvent.Action = plugins.GetAction("status")
							responseEvent.StateMessage = "One of the builds Failed."
							responseEvent.Artifacts["FAILED_BUILDS"] = failedBuilds
							x.events <- e.NewEvent(responseEvent, nil)
							return nil
						}
					}
				} else {
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = err.Error()
					x.events <- e.NewEvent(failedEvent, fmt.Errorf("%s", resp.Status))
					return nil
				}
				timeout += 1
				time.Sleep(1 * time.Second)
				if timeout >= TIMEOUT_LIMIT {
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = err.Error()
					x.events <- e.NewEvent(failedEvent, fmt.Errorf("Timeout: try again and check if builds are taking too long fome reason."))
					return nil
				}
				log.Info("Looping through again and checking statuses")
			}
		}
	}
	return nil
}
