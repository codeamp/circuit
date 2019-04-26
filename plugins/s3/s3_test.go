package s3_test

import (
	"testing"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/s3"
	"github.com/codeamp/circuit/test"
	"github.com/codeamp/transistor"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type TestSuiteS3Extension struct {
	suite.Suite
	transistor *transistor.Transistor
}

func (suite *TestSuiteS3Extension) SetupSuite() {
	var viperConfig = []byte(`
plugins:
  s3:
    workers: 1
`)

	transistor.RegisterPlugin("s3", func() transistor.Plugin {
		return &s3.S3{S3Interfaces: &MockS3Interface{}}
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

func (suite *TestSuiteS3Extension) TestCreateS3ExtSuccess() {
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
	assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteS3Extension) TestDeleteS3ExtSuccess() {
	event := transistor.NewEvent(plugins.GetEventName("project:s3"), transistor.GetAction("create"), nil)
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

func (suite *TestSuiteS3Extension) buildS3ExtArtifacts() []transistor.Artifact {
	return []transistor.Artifact{
		transistor.Artifact{Key: "aws_access_key_id", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_secret_key", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_region", Value: "us-east-1", Secret: false},
		transistor.Artifact{Key: "aws_bucket", Value: "us-east-1-checkr", Secret: false},
		transistor.Artifact{Key: "aws_prefix", Value: "checkr-deploy-test", Secret: false},
	}
}

func (suite *TestSuiteS3Extension) buildS3ExtArtifactsBadCredentials() []transistor.Artifact {
	return []transistor.Artifact{
		transistor.Artifact{Key: "aws_access_key_id", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_secret_key", Value: "", Secret: false},
		transistor.Artifact{Key: "aws_region", Value: "us-east-1", Secret: false},
		transistor.Artifact{Key: "aws_bucket", Value: "us-east-1-checkr", Secret: false},
		transistor.Artifact{Key: "aws_prefix", Value: "checkr-deploy-test/", Secret: false},
	}
}
