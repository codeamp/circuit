package githubstatus

import (
	"encoding/json"
	"errors"
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

type Status struct {
	Url         string    `json:"url"`
	Id          int       `json:"id"`
	State       string    `json:"state"`
	Description string    `json:"description"`
	TargetUrl   string    `json:"target_url"`
	Context     string    `json:"context"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StatusResponse struct {
	State    string   `json:"state"`
	Statuses []Status `json:"statuses"`
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
		"githubstatus:create",
		"githubstatus:update",
	}
}

type Status struct {
	Url         string    `json:"url"`
	Id          int       `json:"id"`
	State       string    `json:"state"`
	Description string    `json:"description"`
	TargetUrl   string    `json:"target_url"`
	Context     string    `json:"context"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type StatusResponse struct {
	State    string   `json:"state"`
	Statuses []Status `json:"statuses"`
}

func getStatus(e transistor.Event) (StatusResponse, error) {
	username, err := e.GetArtifact("username")
	if err != nil {
		return StatusResponse{}, err
	}

	token, err := e.GetArtifact("personal_access_token")
	if err != nil {
		return StatusResponse{}, err
	}

	payload := e.Payload.(plugins.ReleaseExtension)
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", payload.Release.Project.Repository, payload.Release.HeadFeature.Hash), nil)
	if err != nil {
		return StatusResponse{}, err
	}

	client := &http.Client{}
	req.SetBasicAuth(username.String(), token.String())
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return StatusResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StatusResponse{}, err
	}

	if resp.StatusCode == 200 {
		status := StatusResponse{}
		err = json.Unmarshal(respBody, &status)
		if err != nil {
			return StatusResponse{}, err
		}

		return status, nil
	} else {
		return StatusResponse{}, errors.New(fmt.Sprintf("Unhandled status code: %d", resp.StatusCode))
	}
}

func (x *GithubStatus) Process(e transistor.Event) error {
	log.Info("Processing GithubStatus event")

	timeoutInterval := 5
	userTimeoutInterval, err := e.GetArtifact("timeout_interval")
	if err != nil {
<<<<<<< HEAD
		log.Debug(err.Error())
=======
		log.Error(err.Error(), " ", e)
>>>>>>> WIP on Event Refactor
	} else {
		timeoutInterval, err = strconv.Atoi(userTimeoutInterval.String())
		if err != nil {
			log.Error(err.Error())
		}
	}

	timeoutLimit, err := e.GetArtifact("timeout_seconds")
	if err != nil {
		log.Error(err.Error())
		return err
	}

	username, err := e.GetArtifact("username")
	if err != nil {
		log.Error(err.Error())
		return err
	}

	token, err := e.GetArtifact("personal_access_token")
	if err != nil {
		log.Error(err.Error())
		return err
	}

	timeoutLimitInt, err := strconv.Atoi(timeoutLimit.String())
	if err != nil {
		log.Error(err.Error())
		return err
	}

<<<<<<< HEAD
	if e.Matches("plugins.ProjectExtension") {
		event := e.Payload.(plugins.ProjectExtension)

		switch event.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Name), log.Fields{})
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
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Name), log.Fields{})
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
=======
	if e.Matches("githubstatus:create") {
		switch e.PayloadModel {
		case "plugins.ProjectExtension":
			err = x.createProjectExtension(e, username, token)
			if err != nil {
				x.reportFailureEvent(e, err)
			}
		case "plugins.ReleaseExtension":
			err = x.createReleaseExtension(e, username, token, timeoutLimitInt, timeoutInterval)
			if err != nil {
				x.reportFailureEvent(e, err)
>>>>>>> WIP on Events Refactor
			}
		default:
			log.InfoWithFields(fmt.Sprintf("GithubStatus ProjectExtension event not handled: %s", e.Name), log.Fields{})
			return nil
		}
	}

<<<<<<< HEAD
	if e.Matches("plugins.ReleaseExtension") {
		payload := e.Payload.(plugins.ReleaseExtension)
		switch payload.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Name), log.Fields{
				"hash": payload.Release.HeadFeature.Hash,
			})

			timeout := 0
			for {
				var evt transistor.Event
				var breakTheLoop bool

				log.DebugWithFields("Checking statuses", log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})

				status, err := getStatus(e)
				if err != nil {
					log.ErrorWithFields(err.Error(), log.Fields{
						"hash": payload.Release.HeadFeature.Hash,
					})

					payload.State = plugins.GetState("failed")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = err.Error()
					evt = e.NewEvent(payload, nil)
					x.events <- e.NewEvent(payload, err)
					return nil
				}

				if status.State == "success" {
					breakTheLoop = true
					payload.State = plugins.GetState("complete")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "All status checks successful."
					evt = e.NewEvent(payload, nil)
				} else if status.State == "failure" {
					breakTheLoop = true
					payload.State = plugins.GetState("failed")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "One or more status checks failed."
					evt = e.NewEvent(payload, nil)
				} else if timeout >= timeoutLimitInt {
					breakTheLoop = true
					payload.State = plugins.GetState("failed")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = fmt.Sprintf("%d seconds timeout reached", timeoutLimitInt)
					evt = e.NewEvent(payload, nil)
				} else {
					breakTheLoop = false
					payload.State = plugins.GetState("running")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "One or more status checks are running."
					evt = e.NewEvent(payload, nil)
				}

				// Collect artifacts
				for _, _status := range status.Statuses {
					evt.AddArtifact(fmt.Sprintf("%d_%s_target_url", _status.Id, _status.Context), _status.TargetUrl, false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_created_at", _status.Id, _status.Context), _status.CreatedAt.String(), false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_state", _status.Id, _status.Context), _status.State, false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_description", _status.Id, _status.Context), _status.Description, false)
				}

				x.events <- evt

				if breakTheLoop {
					break
				}

				timeout += timeoutInterval
				time.Sleep(time.Duration(timeoutInterval) * time.Second)
=======
	if e.Matches("githubstatus:update") {
		switch e.PayloadModel {
		case "plugins.ReleaseExtension":
			err = x.updateProjectExtension(e, username, token)
			if err != nil {
				x.reportFailureEvent(e, err)
			}
		}
	}
	return nil
}

func (x *GithubStatus) reportFailureEvent(e transistor.Event, err error) {
	event := e.NewEvent(plugins.GetAction("status"), plugins.GetState("failed"), err.Error())
	x.events <- event
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

func (x *GithubStatus) createProjectExtension(e transistor.Event, username transistor.Artifact, token transistor.Artifact) error {
<<<<<<< HEAD
<<<<<<< HEAD
	log.InfoWithFields(fmt.Sprintf("Process GithubStatus project extension event: %s", e.Event()), log.Fields{})

	event := e.NewEvent(plugins.GetEventName("githubstatus"), plugins.GetAction("status"), e.Payload)
=======
	event := e.NewEvent(plugins.GetAction("status"), e.Payload)
>>>>>>> WIP on Event Refactor
=======

	var event transistor.Event
>>>>>>> WIP on GithubStatus
	if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
		event := e.NewEvent(plugins.GetAction("status"), plugins.GetState("complete"), "Successfully installed!")
	} else {
		event := e.NewEvent(plugins.GetAction("status"), plugins.GetState("failed"), err.Error())
	}

	x.events <- event
	return nil
}

func (x *GithubStatus) updateProjectExtension(e transistor.Event, username transistor.Artifact, token transistor.Artifact) error {
	var event transistor.Event
	if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
		event = e.NewEvent(plugins.GetAction("status"), plugins.GetState("complete"), "Successfully updated!")
	} else {
		event = e.NewEvent(plugins.GetAction("status"), plugins.GetState("failed"), err.Error())
	}

	x.events <- event
	return nil
}

<<<<<<< HEAD
func (x *GithubStatus) createReleaseExtension(e transistor.Event, username transistor.Artifact, token transistor.Artifact, timeoutLimitInt int) error {
	log.InfoWithFields(fmt.Sprintf("Process GithubStatus release extension event: %s", e.Event()), log.Fields{})
=======
func getStatus(e transistor.Event) (StatusResponse, error) {
	username, err := e.GetArtifact("username")
	if err != nil {
		return StatusResponse{}, err
	}

	token, err := e.GetArtifact("personal_access_token")
	if err != nil {
		return StatusResponse{}, err
	}

	payload := e.Payload.(plugins.ReleaseExtension)

	log.Info("Payload: ", payload)
	log.Info(fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", payload.Release.Project.Repository, payload.Release.HeadFeature.Hash))
	req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", payload.Release.Project.Repository, payload.Release.HeadFeature.Hash), nil)
	if err != nil {
		return StatusResponse{}, err
	}

	client := &http.Client{}
	req.SetBasicAuth(username.String(), token.String())
	resp, err := client.Do(req)
	defer resp.Body.Close()
	if err != nil {
		return StatusResponse{}, err
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return StatusResponse{}, err
	}

	log.Info("RespBody: ", resp.StatusCode)

	if resp.StatusCode == 200 {
		status := StatusResponse{}
		err = json.Unmarshal(respBody, &status)
		if err != nil {
			return StatusResponse{}, err
		}

		return status, nil
	} else {
		return StatusResponse{}, errors.New(fmt.Sprintf("Unhandled status code: %d", resp.StatusCode))
	}
}

func (x *GithubStatus) createReleaseExtension(e transistor.Event, username transistor.Artifact, token transistor.Artifact, timeoutLimitInt int, timeoutInterval int) error {
	payload := e.Payload.(plugins.ReleaseExtension)

>>>>>>> WIP on Event Refactor
	// get status and check if complete
	client := &http.Client{}

	timeout := 0
	for {
<<<<<<< HEAD
		payload := e.Payload.(plugins.ReleaseExtension)
		req, err := http.NewRequest("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", payload.Release.Project.Repository, payload.Release.HeadFeature.Hash), nil)
		if err != nil {
			return err
		}
		req.SetBasicAuth(username.String(), token.String())

		resp, _ := client.Do(req)
		if resp.StatusCode == 200 {
			// send an event that we're successfully getting data from github status API
			statusEvent := e.NewEvent(plugins.GetEventName("githubstatus"), plugins.GetAction("status"), e.Payload)
			statusEvent.SetState(plugins.GetState("waiting"), "Successfully got build events. Waiting for builds to succeed.")

			x.events <- statusEvent

			combinedStatusBody, err := ioutil.ReadAll(resp.Body)
=======
		if timeout >= timeoutLimitInt {
			err := fmt.Errorf("%d seconds timeout reached", timeoutLimitInt)
			log.ErrorWithFields(err.Error(), log.Fields{
				"hash": payload.Release.HeadFeature.Hash,
			})

			return err
		} else {
			log.DebugWithFields("Checking statuses", log.Fields{
				"hash": payload.Release.HeadFeature.Hash,
			})

			status, err = getStatus(e)
			log.Info(status.State)
>>>>>>> WIP on Event Refactor
			if err != nil {
				return err
			}
			resp.Body.Close()
			// unmarshal into interface
			var statusBodyInterface interface{}
			if err := json.Unmarshal([]byte(combinedStatusBody), &statusBodyInterface); err != nil {
				return err
			} else {
				log.InfoWithFields("got > 1 status objects", log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})
				statuses := statusBodyInterface.(map[string]interface{})["statuses"].([]interface{})
				if len(statuses) > 0 {
					statusEvent = e.NewEvent(plugins.GetEventName("githubstatus"), plugins.GetAction("status"), e.Payload)
					statusEvent.SetState(plugins.GetState("waiting"), "Successfully got build events. Waiting for builds to succeed.")

					for _, status := range statuses {
						// send back artifacts containing the build urls for each status object
						statusEvent.AddArtifact(fmt.Sprintf("%s_target_url", status.(map[string]string)["context"]), status.(map[string]string)["target_url"], false)
						statusEvent.AddArtifact(fmt.Sprintf("%s_state", status.(map[string]string)["context"]), status.(map[string]string)["state"], false)
					}

					x.events <- statusEvent
				}

				if len(statuses) == 0 || statusBodyInterface.(map[string]interface{})["state"].(string) == "success" {
					statusEvent := e.NewEvent(plugins.GetEventName("githubstatus"), plugins.GetAction("status"), e.Payload)
					statusEvent.SetState(plugins.GetState("complete"), "Completed.")

					x.events <- statusEvent
					return nil
				}

				if statusBodyInterface.(map[string]interface{})["state"].(string) == "failure" {
					failedBuilds := ""
					// check which builds failed
					for _, build := range statusBodyInterface.(map[string]interface{})["statuses"].([]interface{}) {
						if build.(map[string]interface{})["state"].(string) == "failure" {
							failedBuilds += fmt.Sprintf(", %s", build.(map[string]interface{})["context"].(string))
						}
					}

					statusEvent = e.NewEvent(plugins.GetEventName("githubstatus"), plugins.GetAction("status"), e.Payload)
					statusEvent.SetState(plugins.GetState("failed"), "One of the builds Failed.")
					statusEvent.AddArtifact("failed_builds", failedBuilds, false)

					x.events <- statusEvent

					// Should return error here and let upstream function handle the error.
					// In this case theres an artifact to pass back. Need to figure out how to handle these better
					// ADB
					return nil
				}
>>>>>>> WIP on Events Refactor
			}
<<<<<<< HEAD
		} else {
			return fmt.Errorf("%s. %s", resp.Status, err.Error())
		}
		timeout++
		time.Sleep(1 * time.Second)
		if timeout >= timeoutLimitInt {
			return fmt.Errorf("Timeout: try again and check if builds are taking too long fome reason.")
		}

		log.Debug("Looping through again and checking statuses")
=======

			var evt transistor.Event
			if status.State == "success" {
				evt = e.NewEvent(plugins.GetAction("status"), plugins.GetState("complete"), "All status checks successful.")
			} else if status.State == "failure" {
				evt = e.NewEvent(plugins.GetAction("status"), plugins.GetState("failed"), "One or more status checks failed.")
			} else {
				evt = e.NewEvent(plugins.GetAction("status"), plugins.GetState("running"), "One or more status checks are running.")
			}

			for _, _status := range status.Statuses {
				evt.AddArtifact(fmt.Sprintf("%d_%s_target_url", _status.Id, _status.Context), _status.TargetUrl, false)
				evt.AddArtifact(fmt.Sprintf("%d_%s_created_at", _status.Id, _status.Context), _status.CreatedAt.String(), false)
				evt.AddArtifact(fmt.Sprintf("%d_%s_state", _status.Id, _status.Context), _status.State, false)
				evt.AddArtifact(fmt.Sprintf("%d_%s_description", _status.Id, _status.Context), _status.Description, false)
			}

			x.events <- evt

			if status.State == "success" || status.State == "failure" {
				return nil
			}
		}

		timeout += timeoutInterval
		time.Sleep(time.Duration(timeoutInterval) * time.Second)
>>>>>>> WIP on Event Refactor
	}
}
