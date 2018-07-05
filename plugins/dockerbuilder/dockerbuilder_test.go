package dockerbuilder_test

import (
	"strings"
	"testing"

	"github.com/codeamp/circuit/plugins"
	_ "github.com/codeamp/circuit/plugins/dockerbuilder"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuite struct {
	suite.Suite
	transistor *transistor.Transistor
}

var viperConfig = []byte(`
plugins:
  dockerbuilder:
    workers: 1
    workdir: "/tmp/dockerbuilder"
`)

func (suite *TestSuite) SetupSuite() {
	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func (suite *TestSuite) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuite) TestDockerBuilder() {
	var e transistor.Event
	var err error
	deploytestHash := "4930db36d9ef6ef4e6a986b6db2e40ec477c7bc9"

	dockerBuildEvent := plugins.ReleaseExtension{
		Release: plugins.Release{
			Project: plugins.Project{
				Slug:       "dockerbuilder",
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

	ev := transistor.NewEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("create"), dockerBuildEvent)
	ev.AddArtifact("USER", "test", false)
	ev.AddArtifact("PASSWORD", "test", false)
	ev.AddArtifact("EMAIL", "test@checkr.com", false)
	ev.AddArtifact("HOST", "0.0.0.0:5000", false)
	ev.AddArtifact("ORG", "testorg", false)
	suite.transistor.Events <- ev

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("status"), 60)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("running"), e.State)

	e, err = suite.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("status"), 600)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	assert.Equal(suite.T(), transistor.GetAction("status"), e.Action)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)

	if e.State == transistor.GetState("failed") {
		assert.FailNow(suite.T(), e.StateMessage)
	}

	image, err := e.GetArtifact("image")
	imagePrefixCheck := strings.HasPrefix(image.String(), "0.0.0.0:5000/testorg/checkr-deploy-test:")

	if err == nil {
		assert.Equal(suite.T(), true, imagePrefixCheck)
	} else {
		assert.FailNow(suite.T(), err.Error())
	}
}

func TestDockerBuilder(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
