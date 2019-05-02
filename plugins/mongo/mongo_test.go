package mongo_test

import (
	"testing"

	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/mongo"
	"github.com/codeamp/circuit/test"
)

type TestSuiteMongoExtension struct {
	suite.Suite
	transistor     *transistor.Transistor
	mongoInterface mongo.Mongoer
}

func (suite *TestSuiteMongoExtension) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  mongo:
    workers: 1
`)

	suite.MongoInterface = &MockMongoInterface{}
	transistor.RegisterPlugin("Mongo", func() transistor.Plugin {
		return &Mongo.Mongo{MongoInterfaces: suite.MongoInterface.New()}
	}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func TestMongoExtension(t *testing.T) {
	suite.Run(t, new(TestSuiteMongoExtension))
}

func (suite *TestSuiteMongoExtension) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuiteMongoExtension) AfterTest(suiteName string, testName string) {
	// suite.MongoInterface = suite.MongoInterface.New()
	log.Warn("AFTERING TEST")
}

func (suite *TestSuiteMongoExtension) TestCreateMongoExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), suite.buildMongoExtPayload())
	event.Artifacts = suite.buildMongoExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:mongo", transistor.GetAction("status"), 30)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteMongoExtension) TestUpdateMongoExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), suite.buildMongoExtPayload())
	event.Artifacts = suite.buildMongoExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:mongo", transistor.GetAction("status"), 30)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteMongoExtension) TestDeleteMongoExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), suite.buildMongoExtPayload())
	event.Artifacts = suite.buildMongoExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:mongo", transistor.GetAction("status"), 30)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteMongoExtension) TestCreateMongoExtFailMultipleInstall() {
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), nil)
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:mongo", transistor.GetAction("status"), 30)
		if err != nil {
			assert.Nil(suite.T(), err, err.Error())
			return
		}

		if e.State != "running" {
			break
		}
	}

	suite.T().Log(e.StateMessage)
	assert.Equal(suite.T(), transistor.GetState("failed"), e.State)
}

func (suite *TestSuiteMongoExtension) buildMongoExtPayload() plugins.ProjectExtension {
	return plugins.ProjectExtension{
		ID: uuid.NewV4().String(),
		Project: plugins.Project{
			ID:         uuid.NewV4().String(),
			Slug:       "checkr-deploy-test",
			Repository: "checkr/deploy-test",
		},
	}
}

func (suite *TestSuiteMongoExtension) buildMongoExtArtifacts() []transistor.Artifact {
	return []transistor.Artifact{}
}

func (suite *TestSuiteMongoExtension) buildMongoExtArtifactsBadCredentials() []transistor.Artifact {
	return []transistor.Artifact{}
}
