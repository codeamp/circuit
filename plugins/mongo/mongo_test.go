package mongo_test

import (
	"testing"

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
	transistor *transistor.Transistor
	MockMongoAtlasClientNamespace
	MockMongoClientNamespace
}

func (suite *TestSuiteMongoExtension) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  mongo:
    workers: 1
`)
	transistor.RegisterPlugin("mongo", func() transistor.Plugin {
		return &mongo.MongoExtension{
			MongoAtlasClientNamespacer: &suite.MockMongoAtlasClientNamespace,
			MongoClientNamespacer:      &suite.MockMongoClientNamespace,
		}
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
	suite.MockMongoAtlasClientNamespace.Clear()
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

func (suite *TestSuiteMongoExtension) TestCreateMongoExtFailure() {
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), nil)
	event.Artifacts = suite.buildEmptyMongoExtArtifacts()
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
	payload := suite.buildMongoExtPayload()
	event := transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), payload)
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

	event = transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("delete"), payload)
	event.Artifacts = suite.buildMongoExtArtifacts()
	suite.transistor.Events <- event

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
	assert.Equal(suite.T(), e.StateMessage, "Successfully Deleted")
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteMongoExtension) TestCreateMongoExtFailMultipleInstall() {
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

	event = transistor.NewEvent(plugins.GetEventName("project:mongo"), transistor.GetAction("create"), suite.buildMongoExtPayload())
	event.Artifacts = suite.buildMongoExtArtifacts()
	suite.transistor.Events <- event

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
	return []transistor.Artifact{
		transistor.Artifact{Key: "mongo_hostname", Value: "data.Hostname", Secret: false},
		transistor.Artifact{Key: "mongo_database_name", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_endpoint", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_public_key", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_private_key", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_project_id", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_timeout", Value: "10", Secret: false},
		transistor.Artifact{Key: "mongo_credentials_check_timeout", Value: "120", Secret: false},
	}
}

func (suite *TestSuiteMongoExtension) buildEmptyMongoExtArtifacts() []transistor.Artifact {
	return []transistor.Artifact{
		transistor.Artifact{Key: "mongo_hostname", Value: "data.Hostname", Secret: false},
		transistor.Artifact{Key: "mongo_database_name", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_endpoint", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_public_key", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_private_key", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_project_id", Value: "payloadSlug", Secret: false},
		transistor.Artifact{Key: "mongo_atlas_api_timeout", Value: "10", Secret: false},
		transistor.Artifact{Key: "mongo_credentials_check_timeout", Value: "120", Secret: false},
	}
}
