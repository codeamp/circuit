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

func init() {
	transistor.RegisterPlugin("githubstatus", func() transistor.Plugin {
		return &GithubStatus{}
	}, plugins.ReleaseExtension{})
}

func (x *GithubStatus) Description() string {
	return "Get status of commit/ build in Github, whether it's LGTM or CircleCi"
}

func (x *GithubStatus) SampleConfig() string {
	return ` `
}

func (x *GithubStatus) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started GithubStatus")
	return nil
}

func (x *GithubStatus) Stop() {
	log.Info("Stopping GithubStatus")
}

func (x *GithubStatus) Subscribe() []string {
	return []string{
		"project:githubstatus:create",
		"project:githubstatus:update",
		"project:githubstatus:delete",
		"release:githubstatus:create",
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
	if err != nil {
		return StatusResponse{}, err
	}
	defer resp.Body.Close()

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

func (x *GithubStatus) Process(e transistor.Event, workerChan chan transistor.Event, workerID string) error {
	log.Debug("Processing GithubStatus event")

	timeoutInterval := 5
	userTimeoutInterval, err := e.GetArtifact("timeout_interval")
	if err != nil {
		log.Error(err.Error())
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

	if e.Matches("project:githubstatus") {
		switch e.Action {
		case transistor.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Event()), log.Fields{})
			if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Successfully installed!")
				return nil
			} else {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}
		case transistor.GetAction("update"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Event()), log.Fields{})
			if _, err := isValidGithubCredentials(username.String(), token.String()); err == nil {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Successfully updated!")
				return nil
			} else {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}
		default:
			log.WarnWithFields(fmt.Sprintf("GithubStatus ProjectExtension event not handled: %s", e.Event()), log.Fields{})
			return nil
		}
	}

	if e.Matches("release:githubstatus") {
		payload := e.Payload.(plugins.ReleaseExtension)

		switch e.Action {
		case transistor.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Event()), log.Fields{
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

					x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
					return nil
				}

				if status.State == "success" {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "All status checks successful.")
				} else if status.State == "failure" {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), "One or more status checks failed.")
				} else if timeout >= timeoutLimitInt {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%d seconds timeout reached", timeoutLimitInt))
				} else {
					breakTheLoop = false
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("running"), "One or more status checks are running.")
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
			}
		}
	}
	return nil
}
