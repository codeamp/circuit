package graphql_resolver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/codeamp/circuit/plugins"
	"github.com/codeamp/circuit/plugins/codeamp"
	graphql_resolver "github.com/codeamp/circuit/plugins/codeamp/graphql"
	"github.com/codeamp/circuit/plugins/codeamp/model"
	"github.com/codeamp/circuit/plugins/dockerbuilder"
	"github.com/codeamp/circuit/plugins/dockerbuilder/dockerbuilder_mock"
	"github.com/codeamp/circuit/test"
	log "github.com/codeamp/logger"
	"github.com/codeamp/transistor"
	"github.com/davecgh/go-spew/spew"

	"github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"

	graphql "github.com/graph-gophers/graphql-go"
)

type ReleaseTestSuite struct {
	suite.Suite
	Resolver   *graphql_resolver.Resolver
	helper     Helper
	transistor *transistor.Transistor
}

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
      host: "postgres"
      port: "5432"
      user: "postgres"
      dbname: "codeamp"
      sslmode: "disable"
      password: ""
    service_address: ":3012"
  dockerbuilder:
    workers: 1
    workdir: "/tmp/dockerbuilder"  
`)

func (suite *ReleaseTestSuite) SetupTest() {
	migrators := []interface{}{
		&model.Extension{},
		&model.SecretValue{},
		&model.Release{},
		&model.User{},
		&model.ServiceDeploymentStrategy{},
		&model.ServiceHealthProbe{},
		&model.ServicePort{},
		&model.ServiceHealthProbeHttpHeader{},
	}

	db, err := test.SetupResolverTest(migrators)
	if err != nil {
		log.Fatal(err.Error())
	}

	suite.transistor, _ = test.SetupPluginTest(viperConfig)
	transistor.RegisterPlugin("dockerbuilder", func() transistor.Plugin {
		return &dockerbuilder.DockerBuilder{
			Socket:   "unix:///var/run/docker.sock",
			Dockerer: dockerbuilder_mock.MockedDocker{},
		}
	}, plugins.ReleaseExtension{}, plugins.ProjectExtension{})
	transistor.RegisterPlugin("codeamp", func() transistor.Plugin {
		return &codeamp.CodeAmp{
			Events: suite.transistor.Events,
		}
	}, plugins.ReleaseExtension{}, plugins.Release{})
	go suite.transistor.Run()

	suite.Resolver = &graphql_resolver.Resolver{DB: db, Events: make(chan transistor.Event, 10)}
	suite.helper.SetResolver(suite.Resolver, "TestRelease")
	suite.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ReleaseTestSuite) TestCreateReleaseSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)

	portOne := int32(9090)
	scheme := "http"
	path := "/healthz"

	headers := []model.HealthProbeHttpHeaderInput{
		{
			Name:  "X-Forwarded-Proto",
			Value: "https",
		},
		{
			Name:  "X-Forwarded-For",
			Value: "www.example.com",
		},
	}

	healthProbe := model.ServiceHealthProbeInput{
		Method:      "http",
		Port:        &portOne,
		Scheme:      &scheme,
		Path:        &path,
		HttpHeaders: &headers,
	}

	readinessProbe := healthProbe
	livenessProbe := healthProbe
	preStopHook := "sleep 15"

	ts.helper.CreateService(ts.T(), projectResolver, nil, &readinessProbe, &livenessProbe, &preStopHook)

	// Make Project Secret
	envID := string(environmentResolver.ID())
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestCreateReleaseSuccess-project",
		Type:          "env",
		Scope:         "project",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Make Global Secret
	secretInput = model.SecretInput{
		Key:           "TestCreateReleaseSuccess-global",
		Type:          "env",
		Scope:         "global",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
}

func (ts *ReleaseTestSuite) TestCreateReleaseSuccessNoTailFeature() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	// Make Project Secret
	envID := string(environmentResolver.ID())
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestCreateReleaseSuccess-project",
		Type:          "env",
		Scope:         "project",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Make Global Secret
	secretInput = model.SecretInput{
		Key:           "TestCreateReleaseSuccess-global",
		Type:          "env",
		Scope:         "global",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Remove all features so as to force CreateRelease to not use a tail feature id
	err = ts.Resolver.DB.Where("project_id = ?", string(projectResolver.ID())).Delete(&model.Release{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
}

func (ts *ReleaseTestSuite) TestCreateReleaseSuccess_FullFlow() {
	log.Info("TestCreateReleaseSuccess_FullFlow")

	// REQUIRED INPUTS INITIALIZATION

	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())
	envID := string(environmentResolver.ID())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secrets for dockerbuilder config
	dockerUser, err := ts.helper.CreateSecretWithInput(&model.SecretInput{
		Key:           "USER",
		Value:         "test",
		Scope:         "extension",
		EnvironmentID: envID,
		IsSecret:      false,
		Type:          "env",
	})
	assert.Nil(ts.T(), err)

	dockerOrg, err := ts.helper.CreateSecretWithInput(&model.SecretInput{
		Key:           "PASSWORD",
		Value:         "test",
		Scope:         "extension",
		EnvironmentID: envID,
		IsSecret:      false,
		Type:          "env",
	})
	assert.Nil(ts.T(), err)

	dockerPassword, err := ts.helper.CreateSecretWithInput(&model.SecretInput{
		Key:           "EMAIL",
		Value:         "test@checkr.com",
		Scope:         "extension",
		EnvironmentID: envID,
		IsSecret:      false,
		Type:          "env",
	})
	assert.Nil(ts.T(), err)

	dockerEmail, err := ts.helper.CreateSecretWithInput(&model.SecretInput{
		Key:           "HOST",
		Value:         "0.0.0.0:5000",
		Scope:         "extension",
		EnvironmentID: envID,
		IsSecret:      false,
		Type:          "env",
	})
	assert.Nil(ts.T(), err)

	dockerHost, err := ts.helper.CreateSecretWithInput(&model.SecretInput{
		Key:           "ORG",
		Value:         "testorg",
		Scope:         "extension",
		EnvironmentID: envID,
		IsSecret:      false,
		Type:          "env",
	})
	assert.Nil(ts.T(), err)

	// Extension - dockerbuilder
	config := postgres.Jsonb{[]byte(fmt.Sprintf(`[
		{"key": "USER", "value": "%s", "allowOverride": false}, 
		{"key": "ORG", "value": "%s", "allowOverride": false}, 
		{"key": "PASSWORD", "value": "%s", "allowOverride": false}, 
		{"key": "EMAIL", "value": "%s", "allowOverride": false}, 
		{"key": "HOST", "value": "%s", "allowOverride": false}
	]`, string(dockerUser.ID()), string(dockerOrg.ID()), string(dockerPassword.ID()), string(dockerEmail.ID()), string(dockerHost.ID())))}

	extensionResolver := ts.helper.CreateExtensionWithInput(ts.T(), &model.ExtensionInput{
		Name:          "dockerbuilder",
		Key:           "dockerbuilder",
		Component:     "",
		EnvironmentID: envID,
		Cacheable:     true,
		Config:        model.JSON{config.RawMessage},
		Type:          "workflow",
	})

	// ProjectExtensions
	projectExtensionResolver, err := ts.helper.CreateProjectExtensionWithConfig(ts.T(), extensionResolver, projectResolver, &[]model.ExtConfig{}, nil)
	assert.Nil(ts.T(), err)

	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	featureID := string(featureResolver.ID())
	projectID := string(projectResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}
	releaseResolver := ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, releaseInput)

	pluginServices := []plugins.Service{}
	pluginSecrets := []plugins.Secret{}

	releaseEvent := plugins.Release{
		IsRollback:  false,
		ID:          string(releaseResolver.ID()),
		Environment: environmentResolver.Key(),
		HeadFeature: plugins.Feature{
			ID:         string(featureResolver.ID()),
			Hash:       featureResolver.Hash(),
			ParentHash: featureResolver.ParentHash(),
			User:       featureResolver.User(),
			Message:    featureResolver.Message(),
			Created:    featureResolver.Created().Time,
		},
		TailFeature: plugins.Feature{
			ID:         string(featureResolver.ID()),
			Hash:       featureResolver.Hash(),
			ParentHash: featureResolver.ParentHash(),
			User:       featureResolver.User(),
			Message:    featureResolver.Message(),
			Created:    featureResolver.Created().Time,
		},
		User: releaseResolver.User().Email(),
		Project: plugins.Project{
			ID:         string(projectResolver.ID()),
			Slug:       projectResolver.Slug(),
			Repository: projectResolver.Repository(),
		},
		Git: plugins.Git{
			Url:           projectResolver.GitUrl(),
			Branch:        "master",
			RsaPrivateKey: projectResolver.RsaPrivateKey(),
		},
		Secrets:  pluginSecrets,
		Services: pluginServices,
	}

	ts.transistor.Events <- transistor.NewEvent("release", "create", releaseEvent)

	// e, err := ts.transistor.GetTestEvent(plugins.GetEventName("release:dockerbuilder"), transistor.GetAction("status"), 60)
	// if err != nil {
	// 	assert.FailNow(ts.T(), err.Error())
	// }
	// assert.Equal(ts.T(), transistor.GetAction("status"), e.Action)
	// assert.Equal(ts.T(), transistor.GetState("running"), e.State)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureDuplicateRelease() {
	spew.Dump("TestCreateReleaseFailureDuplicateRelease")
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Service
	ts.helper.CreateServiceSpec(ts.T(), true)
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	// Make Project Secret
	envID := string(environmentResolver.ID())
	projectID := string(projectResolver.ID())
	secretInput := model.SecretInput{
		Key:           "TestCreateReleaseSuccess-project",
		Type:          "env",
		Scope:         "project",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Make Global Secret
	secretInput = model.SecretInput{
		Key:           "TestCreateReleaseSuccess-global",
		Type:          "env",
		Scope:         "global",
		EnvironmentID: envID,
		ProjectID:     &projectID,
		IsSecret:      false,
	}
	_, err = ts.helper.CreateSecretWithInput(&secretInput)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoAuth() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Release
	var ctx context.Context
	ts.helper.SetContext(ctx)
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)

	ts.helper.SetContext(test.ResolverAuthContext())
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoProjectExtensions() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Delete Project Extension
	err = ts.Resolver.DB.Delete(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoReleaseExtensions() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)
	extensionResolver.DBExtensionResolver.Extension.Type = "once"
	ts.Resolver.DB.Save(&extensionResolver.DBExtensionResolver.Extension)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Release
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureNoPermission() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	// Delete ProjectEnvironment
	err = ts.Resolver.DB.Where("project_id = ? and environment_id = ?", projectResolver.ID(), environmentResolver.ID()).Delete(&model.ProjectEnvironment{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Release
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

// ADB - 7/18/2018
// This error condition needs more work on the CreateRelease side.
// It's not clear how it is supposed to work when comparing secrets and services signatures
//
// func (ts *ReleaseTestSuite) TestCreateReleaseFailureSameSignatures() {
// 	// Environment
// 	environmentResolver := ts.helper.CreateEnvironment(ts.T())

// 	// Project
// 	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
// 	if err != nil {
// 		assert.FailNow(ts.T(), err.Error())
// 	}

// 	// Secret
// 	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

// 	// Extension
// 	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

// 	// Project Extension
// 	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

// 	// Force to set to 'complete' state for testing purposes
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
// 	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
// 	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

// 	// Feature
// 	projectID := string(projectResolver.ID())
// 	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)
// 	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

// 	featureID := string(featureResolver.ID())
// 	releaseInput := &model.ReleaseInput{
// 		HeadFeatureID: featureID,
// 		ProjectID:     projectID,
// 		EnvironmentID: envID,
// 		ForceRebuild:  false,
// 	}

// 	// Release
// 	ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, releaseInput)

// 	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
// 	log.Error(err)
// 	assert.NotNil(ts.T(), err)
// }

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadHeadFeature() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	featureID := "xxxxxxxx-xxxx-Mxxx-Nxxx-xxxxxxxxxxxx"
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadEnvID() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	projectID := string(projectResolver.ID())
	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()

	releaseInput := &model.ReleaseInput{
		HeadFeatureID: test.InvalidUUID,
		ProjectID:     projectID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseFailureBadProjectID() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	envID := projectResolver.DBProjectResolver.Environment.Model.ID.String()
	featureID := string(featureResolver.ID())
	releaseInput := &model.ReleaseInput{
		HeadFeatureID: featureID,
		ProjectID:     test.InvalidUUID,
		EnvironmentID: envID,
		ForceRebuild:  false,
	}

	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Reset state back to blank
	releaseResolver.DBReleaseResolver.Release.State = ""
	releaseResolver.DBReleaseResolver.Release.StateMessage = "Forced Empty via Test"
	ts.Resolver.DB.Save(&releaseResolver.DBReleaseResolver.Release)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, &releaseInput)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadEnvironment() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Reset state back to blank
	releaseResolver.DBReleaseResolver.Release.State = ""
	releaseResolver.DBReleaseResolver.Release.StateMessage = "Forced Empty via Test"
	ts.Resolver.DB.Save(&releaseResolver.DBReleaseResolver.Release)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: test.InvalidUUID,
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadReleaseID() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Reset state back to blank
	releaseResolver.DBReleaseResolver.Release.State = ""
	releaseResolver.DBReleaseResolver.Release.StateMessage = "Forced Empty via Test"
	ts.Resolver.DB.Save(&releaseResolver.DBReleaseResolver.Release)

	// Rollback
	releaseID := test.ValidUUID
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureReleaseInProgress() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID:            &releaseID,
		ProjectID:     string(projectResolver.ID()),
		HeadFeatureID: string(featureResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestCreateReleaseRollbackFailureBadProjectID() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Rollback
	releaseID := string(releaseResolver.ID())
	releaseInput := model.ReleaseInput{
		ID: &releaseID,
	}
	_, err = ts.helper.CreateReleaseWithError(ts.T(), projectResolver, &releaseInput)
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseSuccess() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureNoAuth() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Failure Case - No Authorization
	var ctx context.Context
	_, err = ts.Resolver.StopRelease(ctx, &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureBadProjectExtension() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the project resolver to trigger the error condition
	ts.Resolver.DB.Delete(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)

}

func (ts *ReleaseTestSuite) TestStopReleaseFailureReleaseNotFound() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the release
	err = ts.Resolver.DB.Where("id = ?", string(releaseResolver.ID())).Delete(&model.Release{}).Error
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureBadReleaseExtension() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	// Release Extension
	ts.helper.CreateReleaseExtension(ts.T(), releaseResolver, projectExtensionResolver)

	// Delete the project resolver to trigger the error condition
	ts.Resolver.DB.Delete(&extensionResolver.DBExtensionResolver.Extension)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	assert.NotNil(ts.T(), err)
}

func (ts *ReleaseTestSuite) TestStopReleaseFailureNoReleaseExtensions() {
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ReleaseTestSuite) TestCreateRollbackReleaseSuccess() {
	var e transistor.Event
	// Environment
	environmentResolver := ts.helper.CreateEnvironment(ts.T())

	// Project
	projectResolver, err := ts.helper.CreateProject(ts.T(), environmentResolver)
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	// Secret
	_ = ts.helper.CreateSecret(ts.T(), projectResolver)

	// Extension
	extensionResolver := ts.helper.CreateExtension(ts.T(), environmentResolver)

	// Project Extension
	projectExtensionResolver := ts.helper.CreateProjectExtension(ts.T(), extensionResolver, projectResolver)

	// Force to set to 'complete' state for testing purposes
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.State = "complete"
	projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension.StateMessage = "Forced Completion via Test"
	ts.Resolver.DB.Save(&projectExtensionResolver.DBProjectExtensionResolver.ProjectExtension)

	// Feature
	featureResolver := ts.helper.CreateFeature(ts.T(), projectResolver)

	// ServiceSpec
	ts.helper.CreateServiceSpec(ts.T(), true)

	// Service
	ts.helper.CreateService(ts.T(), projectResolver, nil, nil, nil, nil)

	// Release
	releaseResolver := ts.helper.CreateRelease(ts.T(), featureResolver, projectResolver)
	for len(ts.Resolver.Events) > 0 {
		<-ts.Resolver.Events
	}

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}

	for len(ts.Resolver.Events) > 0 {
		<-ts.Resolver.Events
	}

	releaseID := string(releaseResolver.ID())

	// Rollback the deploy
	releaseResolver = ts.helper.CreateReleaseWithInput(ts.T(), projectResolver, &model.ReleaseInput{
		ID:            &releaseID,
		HeadFeatureID: string(featureResolver.ID()),
		ProjectID:     string(projectResolver.ID()),
		EnvironmentID: string(environmentResolver.ID()),
	})
	e = <-ts.Resolver.Events

	releaseEvent := e.Payload.(plugins.Release)
	assert.Equal(ts.T(), true, releaseEvent.IsRollback)
	assert.Equal(ts.T(), true, releaseResolver.IsRollback())

	_, err = ts.Resolver.StopRelease(test.ResolverAuthContext(), &struct{ ID graphql.ID }{releaseResolver.ID()})
	if err != nil {
		assert.FailNow(ts.T(), err.Error())
	}
}

func (ts *ReleaseTestSuite) TearDownTest() {
	ts.transistor.Stop()
	time.Sleep(1 * time.Second)
	ts.helper.TearDownTest(ts.T())
	ts.Resolver.DB.Close()
}

func TestSuiteReleaseResolver(t *testing.T) {
	suite.Run(t, new(ReleaseTestSuite))
}
