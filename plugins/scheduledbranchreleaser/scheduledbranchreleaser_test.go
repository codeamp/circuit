package scheduledbranchreleaser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/transistor"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gopkg.in/jarcoal/httpmock.v1"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/test"

	"github.com/codeamp/circuit/plugins/codeamp"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/plugins/scheduledbranchreleaser"
)

type TestSuiteScheduledBranchReleaserExtension struct {
	suite.Suite
	transistor *transistor.Transistor
	Resolver   *graphql_resolver.Resolver

	createdModels []interface{}
}

func (suite *TestSuiteScheduledBranchReleaserExtension) SetupSuite() {
	var viperConfig = []byte(`
redis:
  username:
  password:
  server: "redis:6379"
  database: "0"
  pool: "30"
  process: "1"
plugins:
  codeamp:
    workers: 1
    oidc_uri: http://0.0.0.0:5556/dex
    oidc_client_id: example-app
    postgres:
      host: postgres
      port: "5432"
      user: "postgres"
      dbname: "codeamp"
      sslmode: "disable"
      password: ""
    service_address: ":3012"
  scheduledbranchreleaser:
    workers: 1
    workdir: "/tmp/scheduledbranchreleaser"
`)
	transistor.RegisterPlugin("scheduledbranchreleaser", func() transistor.Plugin {
		return &scheduledbranchreleaser.ScheduledBranchReleaser{}
	}, plugins.ProjectExtension{})

	migrators := []interface{}{
		&model.Project{},
		&model.ProjectExtension{},
	}

	db, err := test.SetupResolverTestWithPath("../../configs/circuit.test.yml", migrators)
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
	}
	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	go suite.transistor.Run()

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder("GET", "https://api.github.com/repos/aballman/helloworld-node", httpmock.NewStringResponder(200, "{}"))
}

func TestScheduledBranchReleaserExtension(t *testing.T) {
	suite.Run(t, new(TestSuiteScheduledBranchReleaserExtension))
}

func (suite *TestSuiteScheduledBranchReleaserExtension) TearDownSuite() {
	suite.transistor.Stop()
}

func (suite *TestSuiteScheduledBranchReleaserExtension) TestSBRExtHandlePulseSuccess() {
	resolver := suite.Resolver

	////////////////////////////////////////////////////////////////////////////////
	// Environment
	ctx := context.WithValue(context.Background(), "jwt", model.Claims{
		UserID:      uuid.FromStringOrNil(codeamp.ScheduledDeployUUID).String(),
		Email:       "codeamp@codeamp.com",
		Permissions: []string{"admin"},
	})

	environmentResolver, err := resolver.CreateEnvironment(ctx, &struct{ Environment *model.EnvironmentInput }{
		&model.EnvironmentInput{
			Name:      "test-create-mongo-ext-handle-pulse",
			Key:       "testing",
			IsDefault: false,
			Color:     "green",
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}
	suite.createdModels = append(suite.createdModels, environmentResolver.DBEnvironmentResolver.Environment)

	// Project
	continuousDeploy := false
	bookmarked := false
	environmentID := string(environmentResolver.ID())
	branch := "testing"

	// Create project ignores whichever branch you provide it. Yay.
	// So that means we'll have to do an immediate update of project afterwards to make it be 'not master'
	projectResolver, err := resolver.CreateProject(ctx, &struct{ Project *model.ProjectInput }{
		&model.ProjectInput{
			GitProtocol:      "https",
			GitUrl:           "https://github.com/aballman/helloworld-node.git",
			GitBranch:        &branch,
			Bookmarked:       &bookmarked,
			ContinuousDeploy: &continuousDeploy,
			EnvironmentID:    &environmentID,
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}

	projectID := string(projectResolver.ID())
	suite.createdModels = append(suite.createdModels, projectResolver.DBProjectResolver.Project)
	_, err = resolver.UpdateProject(ctx, &struct{ Project *model.ProjectInput }{
		&model.ProjectInput{
			GitProtocol:      "https",
			GitUrl:           "https://github.com/aballman/helloworld-node.git",
			GitBranch:        &branch,
			Bookmarked:       &bookmarked,
			ContinuousDeploy: &continuousDeploy,
			EnvironmentID:    &environmentID,
			ID:               &projectID,
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}

	// Project Environment (permissions)
	_, err = resolver.UpdateProjectEnvironments(ctx, &struct {
		ProjectEnvironments *model.ProjectEnvironmentsInput
	}{
		&model.ProjectEnvironmentsInput{
			ProjectID: string(projectResolver.ID()),
			Permissions: []model.ProjectEnvironmentInput{
				{
					EnvironmentID: string(environmentResolver.ID()),
					Grant:         true,
				},
			},
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}

	// Secret
	branchSecretResolver, err := resolver.CreateSecret(ctx, &struct{ Secret *model.SecretInput }{
		&model.SecretInput{
			Key:           "BRANCH",
			Value:         "master",
			Scope:         "extension",
			EnvironmentID: string(environmentResolver.ID()),
			Type:          "env",
		},
	})

	scheduleSecretResolver, err := resolver.CreateSecret(ctx, &struct{ Secret *model.SecretInput }{
		&model.SecretInput{
			Key:           "schedule",
			Value:         time.Now().UTC().Format("15:04 -0700 UTC"),
			Scope:         "extension",
			EnvironmentID: string(environmentResolver.ID()),
			Type:          "env",
		},
	})

	// Extension
	extensionConfig := fmt.Sprintf(`[
	   {
	      "key":"BRANCH",
	      "value":"%s",
	      "allowOverride":true
	   },
	   {
	      "key":"SCHEDULE",
	      "value":"%s",
	      "allowOverride":true
	   }
	]`, branchSecretResolver.ID(), scheduleSecretResolver.ID())
	extensionResolver, err := resolver.CreateExtension(&struct{ Extension *model.ExtensionInput }{
		&model.ExtensionInput{
			Name:          "scheduledbranchreleaser",
			Key:           "scheduledbranchreleaser",
			EnvironmentID: string(environmentResolver.ID()),
			Cacheable:     false,
			Config:        model.JSON{RawMessage: json.RawMessage([]byte(extensionConfig))},
			Type:          "once",
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}
	suite.createdModels = append(suite.createdModels, extensionResolver.DBExtensionResolver.Extension)

	testExtResolver, err := resolver.CreateExtension(&struct{ Extension *model.ExtensionInput }{
		&model.ExtensionInput{
			Name:          "workflowtest",
			Key:           "workflowtest",
			EnvironmentID: string(environmentResolver.ID()),
			Cacheable:     false,
			Config:        model.JSON{RawMessage: json.RawMessage([]byte("[]"))},
			Type:          "workflow",
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}
	suite.createdModels = append(suite.createdModels, testExtResolver.DBExtensionResolver.Extension)

	// Service Spec
	serviceSpecResolver, err := resolver.CreateServiceSpec(&struct{ ServiceSpec *model.ServiceSpecInput }{
		&model.ServiceSpecInput{
			Name:                   "test",
			CpuRequest:             "100",
			CpuLimit:               "100",
			MemoryRequest:          "100",
			MemoryLimit:            "100",
			TerminationGracePeriod: "60",
			IsDefault:              true,
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}
	suite.createdModels = append(suite.createdModels, serviceSpecResolver.DBServiceSpecResolver.ServiceSpec)

	// Service
	serviceResolver, err := resolver.CreateService(&struct{ Service *model.ServiceInput }{
		&model.ServiceInput{
			ProjectID:          string(projectResolver.ID()),
			Command:            "/bin/true",
			Name:               "test-service",
			Count:              0,
			Ports:              nil,
			Type:               "general",
			EnvironmentID:      string(environmentResolver.ID()),
			DeploymentStrategy: nil,
			ReadinessProbe:     nil,
			LivenessProbe:      nil,
			PreStopHook:        nil,
		},
	})
	if err != nil {
		assert.FailNow(suite.T(), err.Error())
		return
	}
	suite.createdModels = append(suite.createdModels, serviceResolver.DBServiceResolver.Service)

	// Project Extension
	{
		projectExtensionResolver, err := resolver.CreateProjectExtension(ctx, &struct{ ProjectExtension *model.ProjectExtensionInput }{
			&model.ProjectExtensionInput{
				ProjectID:     string(projectResolver.ID()),
				ExtensionID:   string(extensionResolver.ID()),
				Config:        model.JSON{RawMessage: json.RawMessage([]byte("[]"))},
				CustomConfig:  model.JSON{RawMessage: json.RawMessage([]byte("{}"))},
				EnvironmentID: string(environmentResolver.ID()),
			},
		})
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
			return
		}
		suite.createdModels = append(suite.createdModels, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)
		projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"

		if err := suite.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension).Error; err != nil {
			assert.FailNow(suite.T(), err.Error())
			return
		}
	}

	// Project Extension
	{
		projectExtensionResolver, err := resolver.CreateProjectExtension(ctx, &struct{ ProjectExtension *model.ProjectExtensionInput }{
			&model.ProjectExtensionInput{
				ProjectID:     string(projectResolver.ID()),
				ExtensionID:   string(testExtResolver.ID()),
				Config:        model.JSON{RawMessage: json.RawMessage([]byte("[]"))},
				CustomConfig:  model.JSON{RawMessage: json.RawMessage([]byte("{}"))},
				EnvironmentID: string(environmentResolver.ID()),
			},
		})
		if err != nil {
			assert.FailNow(suite.T(), err.Error())
			return
		}
		suite.createdModels = append(suite.createdModels, projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)
		projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
		if err := suite.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension).Error; err != nil {
			assert.FailNow(suite.T(), err.Error())
			return
		}
	}

	// event := transistor.NewEvent(plugins.GetEventName("heartbeat"), transistor.GetAction("status"), plugins.HeartBeat{Tick: "minute"})
	// suite.transistor.Events <- event

	// e, err := suite.transistor.GetTestEvent("scheduledbranchreleaser:scheduled", transistor.GetAction("status"), 61)
	// if err != nil {
	// 	assert.Nil(suite.T(), err, err.Error())
	// 	return
	// }

	// suite.T().Log(e.StateMessage)
	// assert.Equal(suite.T(), transistor.GetState("complete"), e.State)
}

func (suite *TestSuiteScheduledBranchReleaserExtension) AfterTest(suiteName string, testName string) {
	for _, model := range suite.createdModels {
		suite.Resolver.DB.Unscoped().Delete(model)
	}
	suite.createdModels = make([]interface{}, 0, 10)
}

func (suite *TestSuiteScheduledBranchReleaserExtension) buildExtArtifacts() []transistor.Artifact {
	return []transistor.Artifact{
		{Key: "branch", Value: "master", Secret: false},
		{Key: "schedule", Value: "00:00 -0700 UTC", Secret: false},
	}
}
