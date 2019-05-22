package s3_test

import (
	"fmt"
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/s3"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	uuid "github.com/satori/go.uuid"
)

type TestSuiteS3Extension struct {
	suite.Suite
	transistor  *transistor.Transistor
	s3Interface s3.S3Interfacer
}

func (suite *TestSuiteS3Extension) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  s3:
    workers: 1
`)

	suite.s3Interface = &MockS3Interface{}
	transistor.RegisterPlugin("s3", func() transistor.Plugin {
		return &s3.S3{S3Interfaces: suite.s3Interface.New()}
	}, plugins.ProjectExtension{})

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()
}

func TestS3Extension(t *testing.T) {
	suite.Run(t, new(TestSuiteS3Extension))
}

func (suite *TestSuiteS3Extension) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuiteS3Extension) AfterTest(suiteName string, testName string) {
	suite.s3Interface = suite.s3Interface.New()
}

func (suite *TestSuiteS3Extension) TestCreateS3ExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) TestUpdateS3ExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

	event = transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("update"), suite.buildS3ExtPayload())
	event.Artifacts = e.Artifacts
	suite.transistor.Events <- event

	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

	const expectedArtifactCount = 5
	assert.Equal(suite.T(), expectedArtifactCount, len(e.Artifacts))

	for _, artifact := range e.Artifacts {
		if artifact.Value == nil {
			assert.FailNow(suite.T(), fmt.Sprintf("Artifact with key '%s' was nil!", artifact.Key))
		}
	}
}

func (suite *TestSuiteS3Extension) TestCreateS3ExtFailureDuplicate() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

	event = transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) TestDeleteS3ExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

	event = transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("delete"), suite.buildS3ExtPayload())
	event.Artifacts = suite.buildS3ExtArtifacts()
	suite.transistor.Events <- event

	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) TestCreateS3ExtFailNoArtifacts() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), nil)
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) TestCreateS3ExtFailBadCredentials() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), nil)

	event.Artifacts = suite.buildS3ExtArtifactsBadCredentials()
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) TestCreateS3ExtFailMultipleInstall() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), nil)
	suite.transistor.Events <- event

	var e transistor.Event
	var err error
	for {
		e, err = suite.transistor.GetTestEvent("project:s3", transistor.GetAction("status"), 30)
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

func (suite *TestSuiteS3Extension) buildS3ExtPayload() plugins.ProjectExtension {
	return plugins.ProjectExtension{
		ID: uuid.NewV4().String(),
		Project: plugins.Project{
			ID:         uuid.NewV4().String(),
			Slug:       "checkr-deploy-test",
			Repository: "checkr/deploy-test",
		},
	}
}

func (suite *TestSuiteS3Extension) buildS3ExtArtifacts() []transistor.Artifact {
	return []transistor.Artifact{
		transistor.Artifact{Key: "aws_access_key_id", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_secret_key", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_region", Value: "us-east-1", Secret: false},
		transistor.Artifact{Key: "aws_bucket_prefix", Value: "us-east-1", Secret: false},
		transistor.Artifact{Key: "aws_generated_user_prefix", Value: "codeamp-testing-", Secret: false},
		transistor.Artifact{Key: "aws_user_group_name", Value: "codeamp-testing-", Secret: false},
		transistor.Artifact{Key: "aws_credentials_timeout", Value: "10", Secret: false},
	}
}

func (suite *TestSuiteS3Extension) buildS3ExtArtifactsBadCredentials() []transistor.Artifact {
	return []transistor.Artifact{
		transistor.Artifact{Key: "aws_access_key_id", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_secret_key", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_region", Value: "us-east-1", Secret: false},
		transistor.Artifact{Key: "aws_bucket_prefix", Value: "testing-bucket", Secret: false},
	}
}
