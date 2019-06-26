package drone

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"
	"github.com/machinebox/graphql"

	log "github.com/codeamp/logger"
)

type Drone struct {
	events chan transistor.Event
}

func init() {
	transistor.RegisterPlugin("drone", func() transistor.Plugin {
		return &Drone{}
	}, plugins.ReleaseExtension{})
}

func (x *Drone) Description() string {
	return "Get status of Drone build"
}

func (x *Drone) SampleConfig() string {
	return ` `
}

func (x *Drone) Start(e chan transistor.Event) error {
	x.events = e
	log.Info("Started Drone")
	return nil
}

func (x *Drone) Stop() {
	log.Info("Stopping Drone")
}

func (x *Drone) Subscribe() []string {
	return []string{
		"project:drone:create",
		"project:drone:update",
		"project:drone:delete",
		"release:drone:create",
	}
}

func isValidDroneCredentials(droneUrl string, droneToken string) (bool, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/user", droneUrl), nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", droneToken))
	resp, _ := client.Do(req)
	if resp.StatusCode == 200 {
		return true, nil
	} else {
		return false, fmt.Errorf("Drone credentials are not valid. Please check them and re-install.")
	}
}

type DroneConfig struct {
	Url        string
	Token      string
	Repository string
	Branch     string
	GraphqlUrl string
}

type Build struct {
	Id     int    `json:"id"`
	Number int    `json:"number"`
	Status string `json:"status"`
	Link   string `json:"link"`
	Ref    string `json:"ref"`
}

func getLatestSuccessfulBuild(e transistor.Event, c DroneConfig) (Build, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/repos/%s/builds", c.Url, c.Repository), nil)
	if err != nil {
		return Build{}, err
	}

	client := &http.Client{}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	resp, err := client.Do(req)
	if err != nil {
		return Build{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Build{}, err
	}

	if resp.StatusCode == 200 {
		builds := []Build{}
		err = json.Unmarshal(respBody, &builds)
		if err != nil {
			return Build{}, err
		}

		for _, _build := range builds {
			if _build.Status == "success" && _build.Ref == fmt.Sprintf("refs/heads/%s", c.Branch) {
				return _build, nil
			}
		}

		return Build{}, errors.New("No successful build found")
	} else {
		return Build{}, errors.New(fmt.Sprintf("Unhandled status code: %d", resp.StatusCode))
	}
}

func startBuild(e transistor.Event, c DroneConfig, buildNumber int) (Build, error) {
	req, err := http.NewRequest("POST", fmt.Sprintf("%s/api/repos/%s/builds/%v", c.Url, c.Repository, buildNumber), nil)
	if err != nil {
		return Build{}, err
	}

	client := &http.Client{}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	resp, err := client.Do(req)
	if err != nil {
		return Build{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Build{}, err
	}

	if resp.StatusCode == 200 {
		build := Build{}
		err = json.Unmarshal(respBody, &build)
		if err != nil {
			return Build{}, err
		}
		return build, nil
	} else {
		return Build{}, errors.New(fmt.Sprintf("Unhandled status code: %d", resp.StatusCode))
	}
}

func getBuildStatus(e transistor.Event, c DroneConfig, buildNumber int) (Build, error) {
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/api/repos/%s/builds/%v", c.Url, c.Repository, buildNumber), nil)
	if err != nil {
		return Build{}, err
	}

	client := &http.Client{}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	resp, err := client.Do(req)
	if err != nil {
		return Build{}, err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return Build{}, err
	}

	if resp.StatusCode == 200 {
		build := Build{}
		err = json.Unmarshal(respBody, &build)
		if err != nil {
			return Build{}, err
		}

		return build, nil
	} else {
		return Build{}, errors.New(fmt.Sprintf("Unhandled status code: %d", resp.StatusCode))
	}
}

func (x *Drone) Process(e transistor.Event) error {
	log.Debug("Processing Drone event")

	timeoutInterval := 5
	userTimeoutInterval, err := e.GetArtifact("timeout_interval")
	if err != nil {
		log.Error(err.Error())
	} else {
		timeoutInterval, err = strconv.Atoi(userTimeoutInterval.String())
		if err != nil {
			log.Error(err.Error())
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
			return err
		}
	}

	timeoutLimit, err := e.GetArtifact("timeout_seconds")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	droneUrl, err := e.GetArtifact("drone_url")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	droneToken, err := e.GetArtifact("drone_token")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	droneBranch, err := e.GetArtifact("drone_branch")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	timeoutLimitInt, err := strconv.Atoi(timeoutLimit.String())
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	graphqlUrl, err := e.GetArtifact("graphql_url")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	graphqlToken, err := e.GetArtifact("internal_bearer_token")
	if err != nil {
		log.Error(err.Error())
		x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
		return err
	}

	droneConfig := DroneConfig{
		Url:        droneUrl.String(),
		GraphqlUrl: graphqlUrl.String(),
		Token:      droneToken.String(),
		Branch:     droneBranch.String(),
	}

	if e.Matches("project:drone") {
		switch e.Action {
		case transistor.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process Drone event: %s", e.Event()), log.Fields{})
			if _, err := isValidDroneCredentials(droneUrl.String(), droneToken.String()); err == nil {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Successfully installed!")
				return nil
			} else {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}
		case transistor.GetAction("update"):
			log.InfoWithFields(fmt.Sprintf("Process Drone event: %s", e.Event()), log.Fields{})
			if _, err := isValidDroneCredentials(droneUrl.String(), droneToken.String()); err == nil {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "Successfully updated!")
				return nil
			} else {
				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}
		default:
			log.WarnWithFields(fmt.Sprintf("Drone ProjectExtension event not handled: %s", e.Event()), log.Fields{})
			return nil
		}
	}

	if e.Matches("release:drone") {
		payload := e.Payload.(plugins.ReleaseExtension)

		repository, err := e.GetArtifact("repository")
		if err != nil {
			log.Error(err.Error())
			x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
			return err
		} else if repository.String() == "" {
			droneConfig.Repository = payload.Release.Project.Repository
		} else {
			droneConfig.Repository = repository.String()
		}

		switch e.Action {
		case transistor.GetAction("create"):
			log.InfoWithFields(fmt.Sprintf("Process Drone event: %s", e.Event()), log.Fields{
				"hash": payload.Release.HeadFeature.Hash,
			})

			// Check if sha is deployed in child environment

			graphqlClient := graphql.NewClient(fmt.Sprintf("http://0.0.0.0:3011/query"))

			// make a request
			req := graphql.NewRequest(`
				query{
					project(name: "codeamp/panel"){
				  		id
					}
			  	}
			`)

			// set header fields
			req.Header.Set("Cache-Control", "no-cache")
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", graphqlToken.String()))

			type response struct {
				Project struct {
					Id string
				}
			}

			// run it and capture the response
			var respData response

			if err := graphqlClient.Run(context.Background(), req, &respData); err != nil {
				log.Fatal(err)
			}

			spew.Dump(respData)

			return nil

			// Find latest sucessful build
			successfulbuild, err := getLatestSuccessfulBuild(e, droneConfig)
			if err != nil {
				log.ErrorWithFields(err.Error(), log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})

				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}

			// Create a new build
			build, err := startBuild(e, droneConfig, successfulbuild.Number)
			if err != nil {
				log.ErrorWithFields(err.Error(), log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})

				x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
				return nil
			}

			timeout := 0
			for {
				var evt transistor.Event
				var breakTheLoop bool

				log.DebugWithFields("Checking build status", log.Fields{
					"hash": payload.Release.HeadFeature.Hash,
				})

				status, err := getBuildStatus(e, droneConfig, build.Number)
				if err != nil {
					log.ErrorWithFields(err.Error(), log.Fields{
						"hash": payload.Release.HeadFeature.Hash,
					})

					x.events <- e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), err.Error())
					return nil
				}

				if status.Status == "success" {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("complete"), "All status checks successful.")
				} else if status.Status == "failure" {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), "One or more status checks failed.")
				} else if timeout >= timeoutLimitInt {
					breakTheLoop = true
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("failed"), fmt.Sprintf("%d seconds timeout reached", timeoutLimitInt))
				} else if status.Status == "pending" {
					breakTheLoop = false
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("waiting"), "One or more status checks are pending.")
				} else {
					breakTheLoop = false
					evt = e.NewEvent(transistor.GetAction("status"), transistor.GetState("running"), "One or more status checks are running.")
				}

				evt.AddArtifact("link", fmt.Sprintf("%s/%s/%v", droneConfig.Url, droneConfig.Repository, build.Number), false)

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
