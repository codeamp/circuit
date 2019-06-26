package drone_test

import (
	"fmt"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  drone:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestDrone() {
	//httpmock.Activate()
	httpmock.DeactivateAndReset()

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

	var e transistor.Event
	var err error

	dronePayload := plugins.ReleaseExtension{
		Release: plugins.Release{
			Project: plugins.Project{
				Slug:       "drone",
				Repository: "checkr/deploy-test",
			},
			Git: plugins.Git{
				Url:           "https://github.com/checkr/deploy-test.git",
				Protocol:      "HTTPS",
				Branch:        "master",
				RsaPrivateKey: "",
				RsaPublicKey:  "",
				Workdir:       "/tmp/something",
			},
			HeadFeature: plugins.Feature{
				Hash:       deploytestHash,
				ParentHash: deploytestHash,
				User:       "",
				Message:    "Test",
			},
			Environment: "testing",
		},
	}

	droneBuildsResponse := `
	[
		{
			"id": 166,
			"repo_id": 264,
			"trigger": "sasso",
			"number": 18,
			"status": "failure",
			"event": "push",
			"action": "",
			"link": "https://github.com/codeamp/test/compare/1772a2651530...572f9d89ab5b",
			"timestamp": 0,
			"message": "Update cache_from",
			"before": "1772a265153013de05597330df3901627358bf95",
			"after": "572f9d89ab5bff4e475e69da87aeaa5805f26e82",
			"ref": "refs/heads/drone",
			"source_repo": "",
			"source": "drone",
			"target": "drone",
			"author_login": "",
			"author_name": "Saso Matejina",
			"author_email": "",
			"author_avatar": "https://avatars0.githubusercontent.com/u/8398242?v=4",
			"sender": "sasso",
			"started": 1561486206,
			"finished": 1561486533,
			"created": 1561486203,
			"updated": 1561486206,
			"version": 3
		},
		{
			"id": 165,
			"repo_id": 264,
			"trigger": "sasso",
			"number": 17,
			"status": "success",
			"event": "push",
			"action": "",
			"link": "https://github.com/codeamp/test/compare/1772a2651530...572f9d89ab5b",
			"timestamp": 0,
			"message": "Update cache_from",
			"before": "1772a265153013de05597330df3901627358bf95",
			"after": "572f9d89ab5bff4e475e69da87aeaa5805f26e82",
			"ref": "refs/heads/drone",
			"source_repo": "",
			"source": "drone",
			"target": "drone",
			"author_login": "",
			"author_name": "Saso Matejina",
			"author_email": "",
			"author_avatar": "https://avatars0.githubusercontent.com/u/8398242?v=4",
			"sender": "sasso",
			"started": 1561443975,
			"finished": 1561444262,
			"created": 1561443973,
			"updated": 1561443975,
			"version": 3
		}
	]
	`
	droneSuccessBuildResponse := `
	{
		"id": 166,
		"repo_id": 264,
		"trigger": "sasso",
		"number": 19,
		"status": "success",
		"event": "push",
		"action": "",
		"link": "https://github.com/codeamp/test/compare/1772a2651530...572f9d89ab5b",
		"timestamp": 0,
		"message": "Update cache_from",
		"before": "1772a265153013de05597330df3901627358bf95",
		"after": "572f9d89ab5bff4e475e69da87aeaa5805f26e82",
		"ref": "refs/heads/drone",
		"source_repo": "",
		"source": "drone",
		"target": "drone",
		"author_login": "",
		"author_name": "Saso Matejina",
		"author_email": "",
		"author_avatar": "https://avatars0.githubusercontent.com/u/8398242?v=4",
		"sender": "sasso",
		"started": 1561486206,
		"finished": 1561486533,
		"created": 1561486203,
		"updated": 1561486206,
		"version": 3,
		"stages": [
			{
			"id": 161,
			"repo_id": 264,
			"build_id": 166,
			"number": 1,
			"name": "default",
			"kind": "pipeline",
			"type": "docker",
			"status": "failure",
			"errignore": false,
			"exit_code": 1,
			"machine": "ip-10-101-101-112.ec2.internal",
			"os": "linux",
			"arch": "amd64",
			"started": 1561486206,
			"stopped": 1561486533,
			"created": 1561486203,
			"updated": 1561486533,
			"version": 3,
			"on_success": true,
			"on_failure": false,
			"steps": [
				{
				"id": 631,
				"step_id": 161,
				"number": 1,
				"name": "clone",
				"status": "success",
				"exit_code": 0,
				"started": 1561486206,
				"stopped": 1561486214,
				"version": 4
				},
				{
				"id": 632,
				"step_id": 161,
				"number": 2,
				"name": "build",
				"status": "success",
				"exit_code": 0,
				"started": 1561486214,
				"stopped": 1561486242,
				"version": 4
				},
				{
				"id": 633,
				"step_id": 161,
				"number": 3,
				"name": "test",
				"status": "failure",
				"exit_code": 1,
				"started": 1561486242,
				"stopped": 1561486533,
				"version": 4
				}
			]
			}
		]
	}
	`

	droneRunningBuildResponse := `
	{
		"id": 166,
		"repo_id": 264,
		"trigger": "sasso",
		"number": 19,
		"status": "running",
		"event": "push",
		"action": "",
		"link": "https://github.com/codeamp/test/compare/1772a2651530...572f9d89ab5b",
		"timestamp": 0,
		"message": "Update cache_from",
		"before": "1772a265153013de05597330df3901627358bf95",
		"after": "572f9d89ab5bff4e475e69da87aeaa5805f26e82",
		"ref": "refs/heads/drone",
		"source_repo": "",
		"source": "drone",
		"target": "drone",
		"author_login": "",
		"author_name": "Saso Matejina",
		"author_email": "",
		"author_avatar": "https://avatars0.githubusercontent.com/u/8398242?v=4",
		"sender": "sasso",
		"started": 1561486206,
		"finished": 1561486533,
		"created": 1561486203,
		"updated": 1561486206,
		"version": 3,
		"stages": [
			{
			"id": 161,
			"repo_id": 264,
			"build_id": 166,
			"number": 1,
			"name": "default",
			"kind": "pipeline",
			"type": "docker",
			"status": "failure",
			"errignore": false,
			"exit_code": 1,
			"machine": "ip-10-101-101-112.ec2.internal",
			"os": "linux",
			"arch": "amd64",
			"started": 1561486206,
			"stopped": 1561486533,
			"created": 1561486203,
			"updated": 1561486533,
			"version": 3,
			"on_success": true,
			"on_failure": false,
			"steps": [
				{
				"id": 631,
				"step_id": 161,
				"number": 1,
				"name": "clone",
				"status": "success",
				"exit_code": 0,
				"started": 1561486206,
				"stopped": 1561486214,
				"version": 4
				},
				{
				"id": 632,
				"step_id": 161,
				"number": 2,
				"name": "build",
				"status": "success",
				"exit_code": 0,
				"started": 1561486214,
				"stopped": 1561486242,
				"version": 4
				},
				{
				"id": 633,
				"step_id": 161,
				"number": 3,
				"name": "test",
				"status": "failure",
				"exit_code": 1,
				"started": 1561486242,
				"stopped": 1561486533,
				"version": 4
				}
			]
			}
		]
	}
	`
	droneNewBuildResponse := `
	{"id":167,"repo_id":264,"trigger":"sasso","number":19,"status":"pending","event":"push","action":"","link":"https://github.com/codeamp/test/compare/1772a2651530...572f9d89ab5b","timestamp":0,"message":"Update cache_from","before":"1772a265153013de05597330df3901627358bf95","after":"572f9d89ab5bff4e475e69da87aeaa5805f26e82","ref":"refs/heads/drone","source_repo":"","source":"drone","target":"drone","author_login":"","author_name":"Saso Matejina","author_email":"","author_avatar":"https://avatars0.githubusercontent.com/u/8398242?v=4","sender":"sasso","started":0,"finished":0,"created":1561489899,"updated":1561489899,"version":1}
	`

	droneUrl := "https://drone.test.net"
	repository := "codeamp/test"

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/repos/%s/builds", droneUrl, repository),
		httpmock.NewStringResponder(200, droneBuildsResponse))

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/repos/%s/builds/19", droneUrl, repository),
		httpmock.NewStringResponder(200, droneRunningBuildResponse))

	httpmock.RegisterResponder("POST", fmt.Sprintf("%s/api/repos/%s/builds/17", droneUrl, repository),
		httpmock.NewStringResponder(200, droneNewBuildResponse))

	ev := transistor.NewEvent(plugins.GetEventName("release:drone"), transistor.GetAction("create"), dronePayload)
	ev.AddArtifact("timeout_seconds", "100", false)
	ev.AddArtifact("timeout_interval", "5", false)
	ev.AddArtifact("drone_token", "token", true)
	ev.AddArtifact("drone_branch", "drone", false)
	ev.AddArtifact("drone_url", droneUrl, false)
	ev.AddArtifact("graphql_url", "http://0.0.0.0:3011", false)
	ev.AddArtifact("internal_bearer_token", "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE1MTgwNjMsImV4cCI6MTU5MzA1NDA3MCwiYXVkIjoiY29kZWFtcCIsInN1YiI6IkNnMHdMVE00TlMweU9EQTRPUzB3RWdSdGIyTnIiLCJlbWFpbCI6ImNvZGVhbXBAY29kZWFtcC5vcmciLCJyb2xlIjoibG9jYWwifQ.A_l36r6Nh6-iUTJ2c8OQ0C-4T-ZXmT2CquzultbTk5I", true)
	ev.AddArtifact("repository", repository, false)

	suite.transistor.Events <- ev

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:drone"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error)
		return
	}
	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("running"), e.State)

	httpmock.RegisterResponder("GET", fmt.Sprintf("%s/api/repos/%s/builds/19", droneUrl, repository),
		httpmock.NewStringResponder(200, droneSuccessBuildResponse))

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:drone"), transistor.GetAction("status"), 10)
	if err != nil {
		assert.Nil(suite.T(), err, err.Error)
		return
	}
	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func TestDrone(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
