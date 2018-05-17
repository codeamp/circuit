package githubstatus_test

import (
	"bytes"
	"testing"
	"fmt"

	"github.com/codeamp/circuit/plugins"
	log "github.com/codeamp/logger"
	httpmock "gopkg.in/jarcoal/httpmock.v1"
	"github.com/codeamp/transistor"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  githubstatus:
    workers: 1
`)

func (suite *TestSuite) SetupSuite() {
	viper.SetConfigType("yaml")
	viper.ReadConfig(bytes.NewBuffer(viperConfig))

	config := transistor.Config{
		Plugins:        viper.GetStringMap("plugins"),
		EnabledPlugins: []string{"githubstatus"},
	}

	ag, _ := transistor.NewTestTransistor(config)
	suite.transistor = ag
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestGithubStatus() {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

	var e transistor.Event
	log.SetLogLevel(logrus.DebugLevel)


	githubStatusEvent := plugins.ReleaseExtension{
		Slug:   "githubstatus",
		Action: plugins.GetAction("create"),
		State:  plugins.GetState("waiting"),
		Release: plugins.Release{
			Project: plugins.Project{
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

	githubRunningStatusResponse := `
	{
		"state": "running",
		"statuses": [
			{
				"url": "https://api.github.com/codeamp/circuit/statuses/c29f85a5bbe882d1a2c42be4d13cec8f091c5536",
				"id": 4595890673,
				"state": "running",
				"description": "All good!",
				"target_url": "url1",
				"context": "codeclimate",
				"created_at": "2018-02-12T22:37:48Z",
				"updated_at": "2018-02-12T22:37:48Z"
			},
			{
				"url": "https://api.github.com/codeamp/circuit/statuses/c29f85a5bbe882d1a2c42be4d13cec8f091c5536",
				"id": 4595906284,
				"state": "running",
				"description": "Your tests passed on CircleCI!",
				"target_url": "url2",
				"context": "ci/circleci",
				"created_at": "2018-02-12T22:41:29Z",
				"updated_at": "2018-02-12T22:41:29Z"
			}
		]
	}	
	`

	githubSuccessStatusResponse := `
	{
		"state": "success",
		"statuses": [
			{
				"url": "https://api.github.com/codeamp/circuit/statuses/c29f85a5bbe882d1a2c42be4d13cec8f091c5536",
				"id": 4595890673,
				"state": "success",
				"description": "All good!",
				"target_url": "url1",
				"context": "codeclimate",
				"created_at": "2018-02-12T22:37:48Z",
				"updated_at": "2018-02-12T22:37:48Z"
			},
			{
				"url": "https://api.github.com/codeamp/circuit/statuses/c29f85a5bbe882d1a2c42be4d13cec8f091c5536",
				"id": 4595906284,
				"state": "success",
				"description": "Your tests passed on CircleCI!",
				"target_url": "url2",
				"context": "ci/circleci",
				"created_at": "2018-02-12T22:41:29Z",
				"updated_at": "2018-02-12T22:41:29Z"
			}
		]
	}	
	`

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", githubStatusEvent.Release.Project.Repository, githubStatusEvent.Release.HeadFeature.Hash),
		httpmock.NewStringResponder(200, githubRunningStatusResponse))	

	ev := transistor.NewEvent(githubStatusEvent, nil)
	ev.AddArtifact("timeout_seconds", "100", false)
	ev.AddArtifact("personal_access_token", "test", false)
	ev.AddArtifact("username", "test", false)

	suite.transistor.Events <- ev
	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:githubstatus", 60)
	payload := e.Payload.(plugins.ReleaseExtension)
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("running")), string(payload.State))

	httpmock.RegisterResponder("GET", fmt.Sprintf("https://api.github.com/repos/%s/commits/%s/status", githubStatusEvent.Release.Project.Repository, githubStatusEvent.Release.HeadFeature.Hash),
		httpmock.NewStringResponder(200, githubSuccessStatusResponse))		

	e = suite.transistor.GetTestEvent("plugins.ReleaseExtension:status:githubstatus", 60)
	payload = e.Payload.(plugins.ReleaseExtension)	
	assert.Equal(suite.T(), string(plugins.GetAction("status")), string(payload.Action))
	assert.Equal(suite.T(), string(plugins.GetState("complete")), string(payload.State))
}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
