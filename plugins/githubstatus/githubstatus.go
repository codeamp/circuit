package githubstatus

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
	"errors"

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
	
type Status struct {
	Url string `json:"url"`
	Id int `json:"id"`
	State string `json:"state"`
	Description string `json:"description"`
	TargetUrl string `json:"target_url"`
	Context string `json:"context"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type StatusResponse struct {
	State string `json:"state:"`
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
			}
		default:
			log.InfoWithFields(fmt.Sprintf("GithubStatus ProjectExtension event not handled: %s", e.Name), log.Fields{})			
			return nil
		}
	}

	if e.Matches("plugins.ReleaseExtension") {
		payload := e.Payload.(plugins.ReleaseExtension)
		switch payload.Action {
		case plugins.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process GithubStatus event: %s", e.Name), log.Fields{
				"hash": payload.Release.HeadFeature.Hash,
			})
			// get status and check if complete
			status, err := getStatus(e)
			if err != nil {
				log.ErrorWithFields(err.Error(), log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})
				payload.State = plugins.GetState("failed")
				payload.Action = plugins.GetAction("status")
				payload.StateMessage = err.Error()
				x.events <- e.NewEvent(payload, err)
				return nil
			}

			timeout := 0			
			for {
				evt := e.NewEvent(payload, nil)				
				for _, 	_status := range status.Statuses {
					evt.AddArtifact(fmt.Sprintf("%d_%s_target_url", _status.Id, _status.Context), _status.TargetUrl, false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_created_at", _status.Id, _status.Context), _status.CreatedAt.String(), false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_state", _status.Id, _status.Context), _status.State, false)
					evt.AddArtifact(fmt.Sprintf("%d_%s_description", _status.Id, _status.Context), _status.Description, false)
				}
				
				if status.State == "success" {
					payload.State = plugins.GetState("complete")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "All status checks successful."					
					evt.Payload = payload

					x.events <- evt
					break
				} else if status.State == "failure" {
					payload.State = plugins.GetState("failed")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "One or more status checks failed."
					evt.Payload = payload

					x.events <- evt					
					break
				} else {
					payload.State = plugins.GetState("running")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = "One or more status checks are running."
					evt.Payload = payload

					x.events <- evt
				}

				timeout += timeoutInterval
				time.Sleep(time.Duration(timeoutInterval) * time.Second)
				if timeout >= timeoutLimitInt {
					timeoutErrMsg := fmt.Sprintf("%d seconds timeout reached", timeoutLimitInt)
					log.ErrorWithFields(timeoutErrMsg, log.Fields{
						"hash": payload.Release.HeadFeature.Hash,
					})					
					payload.State = plugins.GetState("failed")
					payload.Action = plugins.GetAction("status")
					payload.StateMessage = timeoutErrMsg
					evt.Payload = payload

					x.events <- evt
					break
				} else {
					log.DebugWithFields("Checking statuses", log.Fields{
						"hash": payload.Release.HeadFeature.Hash,
					})
					status, err = getStatus(e)
					if err != nil {
						log.ErrorWithFields(err.Error(), log.Fields{
							"hash": payload.Release.HeadFeature.Hash,
						})
						payload.State = plugins.GetState("failed")
						payload.Action = plugins.GetAction("status")
						payload.StateMessage = err.Error()
						evt.Payload = payload
	
						x.events <- evt
						break
					}	
				}			
			}
		}
	}
	return nil
}
