package githubstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
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
		"plugins.ProjectExtension:create:githubstatus",
		"plugins.ProjectExtension:update:githubstatus",
		"plugins.ReleaseExtension:create:githubstatus",
	}
}

func isValidGithubCredentials(username string, token string) (bool, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/user"), nil)
	if err != nil {
		return false, err
	}
	req.SetBasicAuth(username, token)
	resp, _ := client.Do(req)
	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, fmt.Errorf("Github credentials are not valid. Please check them and re-install.")
	}
}

func (x *GithubStatus) Process(e transistor.Event) error {
	log.Info("Processing GithubStatus event")
	e.Dump()

	timeoutLimit, err := e.GetArtifact("timeout_seconds")
	if err != nil {
		return err
	}

	timeoutLimitInt, err := strconv.Atoi(timeoutLimit.String())
	if err != nil {
		return err
	}

	username, err := e.GetArtifact("username")
	if err != nil {
		return err
	}

	token, err := e.GetArtifact("personal_access_token")
	if err != nil {
		return err
	}

	if e.Matches("plugins.ProjectExtension") {
		event := e.Payload.(plugins.ProjectExtension)

		switch event.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus project extension event: %s", e.Name), log.Fields{})
			if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
				responseEvent := e.Payload.(plugins.ProjectExtension)
				responseEvent.State = plugins.GetState("complete")
				responseEvent.Action = plugins.GetAction("status")
				responseEvent.StateMessage = "Successfully installed!"
				x.events <- e.NewEvent(responseEvent, nil)
				return nil
			} else {
				failedEvent := e.Payload.(plugins.ProjectExtension)
				failedEvent.State = plugins.GetState("failed")
				failedEvent.Action = plugins.GetAction("status")
				failedEvent.StateMessage = err.Error()
				x.events <- e.NewEvent(failedEvent, err)
				return nil
			}
		case plugins.GetAction("update"):
			if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
				responseEvent := e.Payload.(plugins.ProjectExtension)
				responseEvent.State = plugins.GetState("complete")
				responseEvent.Action = plugins.GetAction("status")
				responseEvent.StateMessage = "Successfully updated!"
				x.events <- e.NewEvent(responseEvent, nil)
				return nil
			} else {
				failedEvent := e.Payload.(plugins.ProjectExtension)
				failedEvent.State = plugins.GetState("failed")
				failedEvent.Action = plugins.GetAction("status")
				failedEvent.StateMessage = err.Error()
				x.events <- e.NewEvent(failedEvent, err)
				return nil
			}
		}
	}

	if e.Matches("plugins.ReleaseExtension") {
		event := e.Payload.(plugins.ReleaseExtension)
		switch event.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus release extension event: %s", e.Name), log.Fields{
				"hash": event.Release.HeadFeature.Hash,
			})
			// get status and check if complete
			client := &http.Client{}

			if err != nil {
				failedEvent := e.Payload.(plugins.ProjectExtension)
				failedEvent.State = plugins.GetState("failed")
				failedEvent.Action = plugins.GetAction("status")
				failedEvent.StateMessage = err.Error()
				x.events <- e.NewEvent(failedEvent, err)
				return nil
			}

			timeout := 0
			for {
				req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", event.Release.Project.Repository, event.Release.HeadFeature.Hash), nil)
				if err != nil {
					log.ErrorWithFields(err.Error(), log.Fields{
						"hash": event.Release.HeadFeature.Hash,
					})
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = err.Error()
					x.events <- e.NewEvent(failedEvent, err)
					return nil
				}
				req.SetBasicAuth(username.String(), token.String())

				resp, _ := client.Do(req)
				if resp.StatusCode == 200 {
					log.InfoWithFields("Successfully got build events. Waiting for builds to succeed.", log.Fields{
						"hash": event.Release.HeadFeature.Hash,
					})
					// send an event that we're successfully getting data from github status API
					statusEvent := e.Payload.(plugins.ReleaseExtension)
					statusEvent.State = plugins.GetState("waiting")
					statusEvent.Action = plugins.GetAction("status")					
					statusEvent.StateMessage = "Successfully got build events. Waiting for builds to succeed."
					x.events <- e.NewEvent(statusEvent, nil)

					combinedStatusBody, err := ioutil.ReadAll(resp.Body)
					if err != nil {
						log.ErrorWithFields(err.Error(), log.Fields{
							"hash": event.Release.HeadFeature.Hash,
						})						
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
						log.InfoWithFields(err.Error(), log.Fields{
							"hash": event.Release.HeadFeature.Hash,
						})						
						failedEvent := e.Payload.(plugins.ReleaseExtension)
						failedEvent.State = plugins.GetState("failed")
						failedEvent.Action = plugins.GetAction("status")
						failedEvent.StateMessage = err.Error()
						x.events <- e.NewEvent(failedEvent, err)
						return nil
					} else {
						log.InfoWithFields("got > 1 status objects", log.Fields{
							"hash": event.Release.HeadFeature.Hash,
						})
						statuses := statusBodyInterface.(map[string]interface{})["statuses"].([]interface{})
						if len(statuses) > 0 {
							statusEvent := e.Payload.(plugins.ReleaseExtension)
							statusEvent.State = plugins.GetState("waiting")
							statusEvent.Action = plugins.GetAction("status")					
							statusEvent.StateMessage = "Successfully got build events. Waiting for builds to succeed."
							newStatusEvent := e.NewEvent(statusEvent, nil)							
							for _, status := range statuses {
								// send back artifacts containing the build urls for each status object
								newStatusEvent.AddArtifact(fmt.Sprintf("%s_target_url", status.(map[string]string)["context"]), status.(map[string]string)["target_url"], false)
								newStatusEvent.AddArtifact(fmt.Sprintf("%s_state", status.(map[string]string)["context"]), status.(map[string]string)["state"], false)
							}

							x.events <- newStatusEvent								
						}

						if len(statuses) == 0 || statusBodyInterface.(map[string]interface{})["state"].(string) == "success" {
							responseEvent := e.Payload.(plugins.ReleaseExtension)
							responseEvent.State = plugins.GetState("complete")
							responseEvent.Action = plugins.GetAction("status")
							responseEvent.StateMessage = "Completed"
							x.events <- e.NewEvent(responseEvent, nil)
							return nil
						}

						if statusBodyInterface.(map[string]interface{})["state"].(string) == "failure" {
							log.InfoWithFields("one of the builds failed", log.Fields{
								"hash": event.Release.HeadFeature.Hash,
							})
							failedBuilds := ""
							// check which builds failed
							for _, build := range statusBodyInterface.(map[string]interface{})["statuses"].([]interface{}) {
								if build.(map[string]interface{})["state"].(string) == "failure" {
									failedBuilds += fmt.Sprintf(", %s", build.(map[string]interface{})["context"].(string))
								}
							}
							responseEvent := e.Payload.(plugins.ReleaseExtension)
							responseEvent.State = plugins.GetState("failed")
							responseEvent.Action = plugins.GetAction("status")
							responseEvent.StateMessage = "One of the builds Failed."

							ev := e.NewEvent(responseEvent, nil)
							ev.AddArtifact("failed_builds", failedBuilds, false)
							x.events <- ev

							return nil
						}
					}
				} else {
					log.ErrorWithFields("failed to get a 200 response", log.Fields{
						"hash": event.Release.HeadFeature.Hash,
					})					
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = err.Error()
					x.events <- e.NewEvent(failedEvent, fmt.Errorf("%s", resp.Status))
					return nil
				}
				timeout++
				time.Sleep(1 * time.Second)
				if timeout >= timeoutLimitInt {
					timeoutErrMsg := fmt.Sprintf("Timeout: try again and check if builds are taking too long for some reason.")
					log.InfoWithFields(timeoutErrMsg, log.Fields{
						"hash": event.Release.HeadFeature.Hash,
					})					
					failedEvent := e.Payload.(plugins.ReleaseExtension)
					failedEvent.State = plugins.GetState("failed")
					failedEvent.Action = plugins.GetAction("status")
					failedEvent.StateMessage = timeoutErrMsg
					x.events <- e.NewEvent(failedEvent, nil)
					return nil
				}
				log.DebugWithFields("Looping through again and checking statuses", log.Fields{
					"hash": event.Release.HeadFeature.Hash,
				})
			}
		}
	}
	return nil
}
